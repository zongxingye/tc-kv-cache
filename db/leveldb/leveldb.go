package leveldb

import (
	"context"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"log"
	"tc-kv-cache/global_mata"
	"tc-kv-cache/iface"
)

type Engine struct {
	ldb         *leveldb.DB
	varindex    int
	indexVarBuf []byte
}

func NewEngine() *Engine {
	//db, err := leveldb.OpenFile(iface.DataDIR, nil)
	db, err := leveldb.RecoverFile(global_mata.DbDir, &opt.Options{
		//BlockCacheCapacity: 1024 * opt.MiB * 16,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &Engine{
		ldb: db,
	}
}
func (s *Engine) Close() {
	if s.ldb != nil {
		s.ldb.Close()
	}
}
func (s *Engine) GetLdbStat() {
	stat := leveldb.DBStats{}
	err := s.ldb.Stats(&stat)
	if err != nil {
		log.Printf("getStats失败  err : %v", err)
		return
	}
	log.Printf("ldbStatus: %+v", stat)
}
func (s *Engine) Init(ctx context.Context) {
	for s.ldb == nil {
	}

	db, err := leveldb.OpenFile(iface.DataDIR, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	iter := db.NewIterator(nil, nil)

	var i = 0
	var kvs []iface.KV
	for iter.Next() {
		i++
		kvs = append(kvs, iface.KV{Key: string(iter.Key()), Value: string(iter.Value())})
		if i == 1000 {
			kvsBak := kvs
			i = 0
			kvs = []iface.KV{}
			go func() {
				s.Batch(ctx, kvsBak)
			}()
		}
	}
	go func() {
		s.Batch(ctx, kvs)
	}()
	iter.Release()
}
func (s *Engine) Init1(ctx context.Context) {
	for s.ldb == nil {
	}

	db, err := leveldb.OpenFile(iface.DataDIR31, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	iter := db.NewIterator(nil, nil)

	var i = 0
	var kvs []iface.KV
	for iter.Next() {
		i++
		kvs = append(kvs, iface.KV{Key: string(iter.Key()), Value: string(iter.Value())})
		if i == 1000 {
			kvsBak := kvs
			i = 0
			kvs = []iface.KV{}

			s.Batch(ctx, kvsBak)

		}
	}

	s.Batch(ctx, kvs)

	iter.Release()
}
func (s *Engine) Init2(ctx context.Context) {
	for s.ldb == nil {
	}

	db, err := leveldb.OpenFile(iface.DataDIR32, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	iter := db.NewIterator(nil, nil)

	var i = 0
	var kvs []iface.KV
	for iter.Next() {
		i++
		kvs = append(kvs, iface.KV{Key: string(iter.Key()), Value: string(iter.Value())})
		if i == 1000 {
			kvsBak := kvs
			i = 0
			kvs = []iface.KV{}

			s.Batch(ctx, kvsBak)

		}
	}

	s.Batch(ctx, kvs)

	iter.Release()
}
func (s *Engine) Init3(ctx context.Context) {
	for s.ldb == nil {
	}

	db, err := leveldb.OpenFile(iface.DataDIR33, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	iter := db.NewIterator(nil, nil)

	var i = 0
	var kvs []iface.KV
	for iter.Next() {
		i++
		kvs = append(kvs, iface.KV{Key: string(iter.Key()), Value: string(iter.Value())})
		if i == 1000 {
			kvsBak := kvs
			i = 0
			kvs = []iface.KV{}

			s.Batch(ctx, kvsBak)

		}
	}

	s.Batch(ctx, kvs)

	iter.Release()
}

func (s Engine) Get(ctx context.Context, k string) (v string, ok bool, err2 error) {
	str, err := s.ldb.Get([]byte(k), nil)
	if err == leveldb.ErrNotFound {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(str), true, nil
}

func (s Engine) Add(ctx context.Context, k string, v string) error {
	return s.ldb.Put([]byte(k), []byte(v), nil)
}

func (s Engine) Del(ctx context.Context, k string) error {
	return s.ldb.Delete([]byte(k), nil)
}

//
//// 不存在返回空的kv对象
//func (s Engine) List(ctx context.Context, ks []string) ([]iface.KV, error) {
//	if len(ks) == 0 {
//		return nil, nil
//	}
//	vals := make([]iface.KV, len(ks))
//	it := s.ldb.NewIterator(nil, nil)
//	defer it.Release()
//	for i := range ks {
//		key := []byte(ks[i])
//		it.Seek(key)
//		if it.Valid() {
//			rk := it.Key()
//			if bytes.Equal(rk, key) {
//				vals[i] = iface.KV{
//					Key:   ks[i],
//					Value: string(it.Value()),
//				}
//			} else {
//				vals[i] = iface.KV{}
//			}
//			continue
//		}
//	}
//	return vals, nil
//}

//不存在不反这个字段
//func (s Engine) List(ctx context.Context, ks []string) ([]iface.KV, error) {
//	if len(ks) == 0 {
//		return nil, nil
//	}
//	vals := make([]iface.KV,0, len(ks))
//	it := s.ldb.NewIterator(nil, nil)
//	defer it.Release()
//	for i := range ks {
//		key := []byte(ks[i])
//		it.Seek(key)
//		if it.Valid() {
//			rk := it.Key()
//			if bytes.Equal(rk, key) {
//				vals = append(vals, iface.KV{
//					Key:   ks[i],
//					Value: string(it.Value()),
//				})
//			}
//			continue
//		}
//	}
//	return vals, nil
//}
// 用for get 试一下
func (s Engine) List(ctx context.Context, ks []string) ([]iface.KV, error) {
	if len(ks) == 0 {
		return nil, nil
	}
	vals := make([]iface.KV, 0, len(ks))
	//it := s.ldb.NewIterator(nil, nil)
	//defer it.Release()
	for i := range ks {
		key := []byte(ks[i])
		//it.Seek(key)
		//if it.Valid() {
		//	rk := it.Key()
		//	if bytes.Equal(rk, key) {
		//		vals = append(vals, iface.KV{
		//			Key:   ks[i],
		//			Value: string(it.Value()),
		//		})
		//	}
		//	continue
		//}
		val, err := s.ldb.Get(key, nil)
		if err == leveldb.ErrNotFound {
			continue
		}
		if err != nil {
			continue
		}
		vals = append(vals, iface.KV{
			Key:   ks[i],
			Value: string(val),
		})
	}
	return vals, nil
}

func (s Engine) Batch(ctx context.Context, kvs []iface.KV) error {
	if len(kvs) == 0 {
		return nil
	}
	batch := new(leveldb.Batch)
	for i := range kvs {
		batch.Put([]byte(kvs[i].Key), []byte(kvs[i].Value))
	}

	return s.ldb.Write(batch, nil)
}

func (s Engine) ZAdd(ctx context.Context, key string, sv iface.SV) error {
	if key == "" {
		return nil
	}
	var num int64 // num的作用是为了增加key对应zset的size，todo需要完成
	tbatch := new(leveldb.Batch)
	if n, err := s.zSetItem(tbatch, []byte(key), int64(sv.Score), []byte(sv.Value)); err != nil {
		return err
	} else if n == 0 {
		//add new
		num++
	}
	// todo 增加zset的size
	// 将tbatch写入ldb
	return s.ldb.Write(tbatch, nil)
}

func (s Engine) ZRange(ctx context.Context, key string, sr iface.ScRange) ([]iface.SV, error) {

	if key == "" {
		return nil, nil
	}
	res, err := s.zRange([]byte(key), int64(sr.MinScore), int64(sr.MaxScore))
	sv := make([]iface.SV, len(res))
	for i := range res {
		sv[i] = iface.SV{
			Score: int(res[i].Score),
			Value: string(res[i].Member),
		}
	}
	return sv, err
}

func (s Engine) ZRmv(ctx context.Context, key string, value string) error {
	if key == "" || value == "" {
		return nil
	}

	return s.ZRem([]byte(key), []byte(value))
}
