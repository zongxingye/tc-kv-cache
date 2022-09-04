package global_mata

import "flag"

var (
	IsLeader int64
)

var (
	DbDir string
)

func init() {
	flag.StringVar(&DbDir, "dbdir", "test/local/data", "db的目录地址")
}