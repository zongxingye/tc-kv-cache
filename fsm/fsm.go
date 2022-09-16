package fsm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"tc-kv-cache/db/leveldb"
	"tc-kv-cache/iface"

	"github.com/hashicorp/raft"
)

type Fsm struct {
	DataBase database
}

func NewFsm() *Fsm {
	fsm := &Fsm{
		DataBase: NewDatabase(),
	}
	return fsm
}

func (f *Fsm) Apply(l *raft.Log) interface{} {
	fmt.Println("apply data:", string(l.Data))
	data := strings.Split(string(l.Data), "@")
	op := data[0]
	log.Println("op为:",op)
	if op == "set" {
		key := data[1]
		value := data[2]
		f.DataBase.Set(key, value)
		if err := f.DataBase.Engine.Add(context.TODO(),key,value);err != nil {
			log.Println("set失败，：",err)
		}
		return nil
	}
	if op == "del" {
		key := data[1]
		if err := f.DataBase.Engine.Del(context.TODO(),key); err != nil {
			log.Println("del失败，：",err)
		}
	return nil
	}
	if op == "zadd" {
		key := data[1]
		svStr := data[2]
		sv := iface.SV{}
		json.Unmarshal([]byte(svStr),&sv)
		if err := f.DataBase.Engine.ZAdd(context.Background(), key, sv);err != nil {
			log.Println("zadd失败，：",err)
		}
		return nil
	}
	if op == "batch" {
		reqStr := data[1]
		log.Println("打印一下reqstr",reqStr)
		req := []iface.KV{}
		err := json.Unmarshal([]byte(reqStr),&req)
		if err != nil {
			log.Println("Unmarshal, failed ：",err)
			return err
		}
		log.Println("batch, unmarshal ：",req)
		if err := f.DataBase.Engine.Batch(context.Background(), req); err != nil {
			log.Println("batch失败，：",err)

		}

	}
	if op == "zrmv"{
		key := data[1]
		value := data[2]
		if err := f.DataBase.Engine.ZRmv(context.Background(), key,value); err != nil {
			log.Println("zrmv失败，：",err)
		}
	}


	return nil
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &f.DataBase, nil
}

func (f *Fsm) Restore(io.ReadCloser) error {
	return nil
}

type database struct {
	Data map[string]string
	Engine *leveldb.Engine
	mu   sync.Mutex
}

func NewDatabase() database {
	return database{
		Data: make(map[string]string),
		Engine: leveldb.NewEngine(),
	}
}

func (d *database) Get(key string) string {
	d.mu.Lock()
	value := d.Data[key]
	d.mu.Unlock()
	return value
}

func (d *database) Set(key, value string) {
	d.mu.Lock()
	d.Data[key] = value
	d.mu.Unlock()
}

func (d *database) Persist(sink raft.SnapshotSink) error {
	d.mu.Lock()
	data, err := json.Marshal(d.Data)
	d.mu.Unlock()
	if err != nil {
		return err
	}
	sink.Write(data)
	sink.Close()
	return nil
}

func (d *database) Release() {}
