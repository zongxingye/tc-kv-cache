package rosedb

import (
	"context"
	"errors"
	"github.com/roseduan/rosedb"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"tc-kv-cache/iface"
)

type Engine struct {
	rosedb *rosedb.RoseDB
	//varindex    int
	//indexVarBuf []byte
}

func NewEngine() *Engine {
	cfg := rosedb.DefaultConfig()
	cfg.DirPath = iface.RoseDBDIR
	//cfg.RwMethod = storage.MMap
	db, err := rosedb.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &Engine{
		rosedb: db,
	}
}

var _ iface.CURD = (*Engine)(nil)

func (s Engine) Init(ctx context.Context) {
	db, err := leveldb.OpenFile(iface.DataDIR, nil)
	defer db.Close()
	if err != nil {
		log.Println(err)
		return
	}
	iter := db.NewIterator(nil, nil)

	var i = 0
	//var kvs []string
	for iter.Next() {
		i++
		//kvs = append(kvs, string(iter.Key()), string(iter.Value()))
		s.rosedb.Set(iter.Key(), iter.Value())
	}

	iter.Release()
}

func (s Engine) Get(ctx context.Context, k string) (v string, ok bool, err2 error) {
	err := s.rosedb.Get(k, &v)

	if err != nil {
		return "", false, err
	} else {
		return v, true, nil
	}
}

func (s Engine) Add(ctx context.Context, k string, v string) error {
	return s.rosedb.Set(k, v)
}

func (s Engine) Del(ctx context.Context, k string) error {
	return s.rosedb.Remove(k)
}

//已验证 mget v == nil 和 v == ""
func (s Engine) List(ctx context.Context, ks []string) ([]iface.KV, error) {
	if len(ks) == 0 {
		return nil, nil
	}
	res := make([]iface.KV, 0, len(ks))
	for i := range ks {
		val := ""
		err := s.rosedb.Get(ks[i], &val)
		if err != nil {
			continue
		}
		res = append(res, iface.KV{
			Key:   ks[i],
			Value: val,
		})
	}
	return res, nil
}

func (s Engine) Batch(ctx context.Context, kvs []iface.KV) error {
	if len(kvs) == 0 {
		return nil
	}

	//var kvsSet []string
	for _, kv := range kvs {
		err := s.rosedb.Set(kv.Key, kv.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s Engine) ZAdd(ctx context.Context, key string, sv iface.SV) error {
	if key == "" {
		return nil
	}
	return s.rosedb.ZAdd([]byte(key), float64(sv.Score), []byte(sv.Value))

}

func (s Engine) ZRange(ctx context.Context, key string, sr iface.ScRange) ([]iface.SV, error) {
	sv := make([]iface.SV, 0)
	if key == "" {
		return nil, nil
	}

	res := s.rosedb.ZScoreRange([]byte(key), float64(sr.MinScore), float64(sr.MaxScore))
	if len(res) == 0 {
		return sv, nil
	}
	if len(res)%2 != 0 {
		return sv, errors.New("res长度不为偶数")
	}
	for i := 0; i < len(res)-1; {
		member := res[i].(string)
		score := res[i+1].(float64)
		i = i + 2
		sv = append(sv, iface.SV{
			Score: int(score),
			Value: member,
		})
	}
	return sv, nil
}

func (s Engine) ZRmv(ctx context.Context, key string, value string) error {
	if key == "" || value == "" {
		return nil
	}

	ok, err := s.rosedb.ZRem([]byte(key), []byte(value))
	if ok {
		return nil
	}
	return err
}
