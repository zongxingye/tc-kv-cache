package global_mata

import "flag"

var (
	IsLeader int64
)

var (
	DbDir string
	HttpAddr    string
)
// 维护一个全局的rpc地址映射到http地址的关系
var AddrMap map[string]string
func init() {
	flag.StringVar(&DbDir, "dbdir", "/sdata", "db的目录地址")
	AddrMap = make(map[string]string)
}