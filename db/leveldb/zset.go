package leveldb

import (
	"encoding/binary"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	//"github.com/syndtr/goleveldb/leveldb/util"
)

type ScorePair struct {
	Score  int64
	Member []byte
}

const index = 0 // default zset with db0
// For zset const.
const (
	MinScore     int64 = -1<<63 + 1
	MaxScore     int64 = 1<<63 - 1
	InvalidScore int64 = -1 << 63

	AggregateSum byte = 0
	AggregateMin byte = 1
	AggregateMax byte = 2
)

const (
	ZSetType   byte = 6
	ZScoreType byte = 8
)

const (
	zsetNScoreSep   byte = '<'
	zsetPScoreSep   byte = zsetNScoreSep + 1
	zsetStartMemSep byte = ':'

	//zsetStartMemSep byte = ':'
	zsetStopMemSep byte = zsetStartMemSep + 1
)

func (s *Engine) setIndex(index int) {
	s.varindex = index
	// the most size for varint is 10 bytes
	buf := make([]byte, 10)
	n := binary.PutUvarint(buf, uint64(index))

	s.indexVarBuf = buf[0:n]
}
func (s Engine) zEncodeSetKey(key []byte, member []byte) []byte {
	buf := make([]byte, len(key)+len(member)+4+len(s.indexVarBuf))

	pos := copy(buf, s.indexVarBuf)

	buf[pos] = ZSetType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = zsetStartMemSep
	pos++

	copy(buf[pos:], member)

	return buf
}
func (s Engine) zEncodeScoreKey(key []byte, member []byte, score int64) []byte {
	buf := make([]byte, len(key)+len(member)+13+len(s.indexVarBuf))

	pos := copy(buf, s.indexVarBuf)

	buf[pos] = ZScoreType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

	copy(buf[pos:], key)
	pos += len(key)

	if score < 0 {
		buf[pos] = zsetNScoreSep
	} else {
		buf[pos] = zsetPScoreSep
	}

	pos++
	binary.BigEndian.PutUint64(buf[pos:], uint64(score))
	pos += 8

	buf[pos] = zsetStartMemSep
	pos++

	copy(buf[pos:], member)
	return buf
}
func (s *Engine) zDecodeScoreKey(ek []byte) (key []byte, member []byte, score int64, err error) {
	pos := 0
	//pos, err = db.checkKeyIndex(ek)
	//if err != nil {
	//	return
	//}

	if pos+1 > len(ek) || ek[pos] != ZScoreType {
		err = errors.New("error socre key")
		return
	}
	pos++

	if pos+2 > len(ek) {
		err = errors.New("error socre key")
		return
	}
	keyLen := int(binary.BigEndian.Uint16(ek[pos:]))
	pos += 2

	if keyLen+pos > len(ek) {
		err = errors.New("error socre key")
		return
	}

	key = ek[pos : pos+keyLen]
	pos += keyLen

	if pos+10 > len(ek) {
		err = errors.New("error socre key")
		return
	}

	if (ek[pos] != zsetNScoreSep) && (ek[pos] != zsetPScoreSep) {
		err = errors.New("error socre key")
		return
	}
	pos++

	score = int64(binary.BigEndian.Uint64(ek[pos:]))
	pos += 8

	if ek[pos] != zsetStartMemSep {
		err = errors.New("error socre key")
		return
	}

	pos++

	member = append([]byte{}, ek[pos:]...)
	return
}

// zadd的base method
func (s *Engine) zSetItem(t *leveldb.Batch, key []byte, score int64, member []byte) (int64, error) {
	if score <= MinScore || score >= MaxScore {
		return 0, errors.New("score over flow")
	}

	var exists int64
	ek := s.zEncodeSetKey(key, member)
	// 先尝试获取
	if v, err := s.ldb.Get(ek, nil); err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return 0, err
	} else if v != nil { // 获取到了就是更新了
		exists = 1

		sint, err := Int64(v, err)
		if err != nil {
			return 0, err
		}

		sk := s.zEncodeScoreKey(key, member, sint)
		t.Delete(sk)
	}

	t.Put(ek, PutInt64(score))

	sk := s.zEncodeScoreKey(key, member, score)
	t.Put(sk, []byte{})

	// debug
	//skval,err :=s.ldb.Get(sk,nil)
	//log.Println(skval,err)
	return exists, nil
}

// Int64 gets 64 integer with the little endian format.
func Int64(v []byte, err error) (int64, error) {
	if err != nil {
		return 0, err
	} else if v == nil || len(v) == 0 {
		return 0, nil
	} else if len(v) != 8 {
		return 0, errors.New("invalid integer")
	}

	return int64(binary.LittleEndian.Uint64(v)), nil
}

// PutInt64 puts the 64 integer.
func PutInt64(v int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return b
}

func (s *Engine) zRange(key []byte, min int64, max int64) ([]ScorePair, error) {
	//校验先不谢
	//if len(key) > MaxKeySize {
	//	return nil, errKeySize
	//}

	v := make([]ScorePair, 0)
	minKey := s.zEncodeStartScoreKey(key, min)
	maxKey := s.zEncodeStopScoreKey(key, max)
	//it := s.ldb.NewIterator(nil,nil)
	it := s.ldb.NewIterator(&util.Range{
		Start: minKey,
		Limit: maxKey,
	}, nil)
	for ok := it.Seek(minKey); ok; ok = it.Next() {

		// Use key/value.
		if it.Valid() {
			_, m, sint, err := s.zDecodeScoreKey(it.Key())
			// 判断是否大于maxkey
			if sint > max {
				break
			}
			if err != nil {
				continue
			}

			v = append(v, ScorePair{Member: m, Score: sint})
		}
	}
	// it :=s.ldb.NewIterator(&util.Range{
	//	 Start: minKey,
	//	 Limit: maxKey,
	// },nil)
	//for ; it.Valid(); it.Next() {
	//	_, m, sint, err := s.zDecodeScoreKey(it.Key())
	//	//may be we will check key equal?
	//	if err != nil {
	//		continue
	//	}
	//
	//	v = append(v, ScorePair{Member: m, Score: sint})
	//}
	it.Release()

	return v, nil
}

func (s *Engine) zEncodeStartScoreKey(key []byte, score int64) []byte {
	return s.zEncodeScoreKey(key, nil, score)
}
func (s *Engine) zEncodeStopScoreKey(key []byte, score int64) []byte {
	k := s.zEncodeScoreKey(key, nil, score)
	k[len(k)-1] = zsetStopMemSep
	return k
}

// ZRem removes  a member
func (s *Engine) ZRem(key []byte, members []byte) error {
	if len(members) == 0 {
		return nil
	}

	t := new(leveldb.Batch)

	// 校验暂时不做
	//if err := checkZSetKMSize(key, members[i]); err != nil {
	//	return 0, err
	//}

	if _, err := s.zDelItem(t, key, members, false); err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return err
	} else if errors.Is(err, leveldb.ErrNotFound) {
		return nil
	}

	return s.ldb.Write(t, nil)

}

func (s *Engine) zDelItem(t *leveldb.Batch, key []byte, member []byte, skipDelScore bool) (int64, error) {
	ek := s.zEncodeSetKey(key, member)
	if v, err := s.ldb.Get(ek, nil); err != nil {
		return 0, err
	} else if v == nil {
		//not exists
		return 0, nil
	} else {
		//exists
		if !skipDelScore {
			//we must del score
			sint, err := Int64(v, err)
			if err != nil {
				return 0, err
			}
			sk := s.zEncodeScoreKey(key, member, sint)
			t.Delete(sk)
		}
	}

	t.Delete(ek)

	return 1, nil
}
