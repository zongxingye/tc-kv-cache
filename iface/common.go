package iface

import (
	"context"
	"errors"
	"hash/crc32"
)

var ErrNotFound = errors.New("404, not found")

//var DataDIR = "data"
var DataDIR = "test/local/data"
var LdbDIR = "test/local/data"
var RoseDBDIR = "test/rosedb"

// prod path
//var DataDIR = "/data"
//var LdbDIR = "/usr/local/data"
//var RoseDBDIR = "/usr/local/data"

//var DataDIR = "/Users/pengshuaifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"
//var LdbDIR = "/Users/pengshuaifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"

//var DataDIR = "/Users/lidifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"
//var LdbDIR = "/Users/lidifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"
//var RoseDBDIR = "/Users/lidifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"

type CURD interface {
	Init(ctx context.Context)
	Get(ctx context.Context, k string) (string, bool, error)
	Add(ctx context.Context, k string, v string) error
	Del(ctx context.Context, k string) error
	List(ctx context.Context, ks []string) ([]KV, error)
	Batch(ctx context.Context, kvs []KV) error
	ZAdd(ctx context.Context, key string, sv SV) error
	ZRange(ctx context.Context, key string, sr ScRange) ([]SV, error)
	ZRmv(ctx context.Context, key string, value string) error
}

type UpdateCluster struct {
	Index int `json:"index"`
	Hosts []string `json:"hosts"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SV struct {
	Score int    `json:"score"`
	Value string `json:"value"`
}

type ScRange struct {
	MinScore int `json:"min_score"`
	MaxScore int `json:"max_score"`
}

func CRC32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}
type TellBody struct {
	Key string `json:"key"`
	Val interface{} `json:"val"`
}