package iface

import (
	"context"
	"errors"
	"hash/crc32"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var ErrNotFound = errors.New("404, not found")

//var DataDIR3 = "data"
//var DataDIR = "test/local/data"
//var LdbDIR = "test/local/data"
//var RoseDBDIR = "test/rosedb"

//prod path
var DataDIR = "/data"
var DataDIR31 = "/data2/data1"
var DataDIR32 = "/data2/data2"
var DataDIR33 = "/data2/data3"
var ClusterDIr = "/node/"


var LdbDIR = "/usr/local/data"
var RoseDBDIR = "/usr/local/data"

//var DataDIR = "/Users/pengshuaifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"
//var LdbDIR = "/Users/pengshuaifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"

//var DataDIR = "/Users/lidifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"
//var LdbDIR = "/Users/lidifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"
//var RoseDBDIR = "/Users/lidifei/Code/go/my-group/tcmatch/tc-memory/wecc/data"
//var DataDIR31 = "data"
//var DataDIR32 = "data"
//var DataDIR33 = "data"
//var ClusterDIr = "node/"
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
// 获取指定文件夹下的所有txt文件
func GetAllFiles(dirPth string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			GetAllFiles(dirPth + PthSep + fi.Name())
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), ".txt")
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := GetAllFiles(table)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		log.Println(err)
		return false
	}
	return true
}
