//
// fasthttp.go
// Copyright (C) 2018 YanMing <yming0221@gmail.com>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"flag"
	fasthttp2 "github.com/valyala/fasthttp"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tc-kv-cache/server/fasthttp"
)

//var (
//	listen   string
//	raftdir  string
//	raftbind string
//	nodeID   string
//	join     string
//)

//func init() {
//	flag.StringVar(&listen, "listen", ":5379", "server listen address")
//	flag.StringVar(&raftdir, "raftdir", "./", "raft data directory")
//	flag.StringVar(&raftbind, "raftbind", ":15379", "raft bus transport bind address")
//	flag.StringVar(&nodeID, "id", "", "node id")
//	flag.StringVar(&join, "join", "", "join to already exist cluster")
//}

var (
	httpAddr    string
	raftAddr    string
	raftId      string
	raftCluster string
	raftDir     string
)

func init() {
	flag.StringVar(&httpAddr, "http_addr", "0.0.0.0:8080", "http listen addr")
	flag.StringVar(&raftAddr, "raft_addr", "127.0.0.1:7000", "raft listen addr")
	flag.StringVar(&raftId, "raft_id", "1", "raft id")
	flag.StringVar(&raftCluster, "raft_cluster", "1/127.0.0.1:7000,2/127.0.0.1:8000,3/127.0.0.1:9000", "cluster info")
}
func main()  {
	server :=  fasthttp.NewFastHTTPSr(nil,nil)
	log.Println("运行http服务开始：")
	go log.Fatal(fasthttp2.ListenAndServe(httpAddr, server.Router.Handler))
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	//启动服务
	//go app.Run()

	<-quitCh
}
//func main() {
//	// 解析参数
//	//flag.Parse()
//	//if httpAddr == "" || raftAddr == "" || raftId == "" || raftCluster == "" {
//	//	fmt.Println("config error")
//	//	os.Exit(1)
//	//	return
//	//}
//	//raftDir := "node/raft_" + raftId
//	//os.MkdirAll(raftDir, 0700)
//	//
//	//// 初始化raft
//	//myRaft, fm, err := myraft.NewMyRaft(raftAddr, raftId, raftDir)
//	//if err != nil {
//	//	fmt.Println("NewMyRaft error ", err)
//	//	os.Exit(1)
//	//	return
//	//}
//
//	// 启动raft
//	//myraft.Bootstrap(myRaft, raftId, raftAddr, raftCluster)
//
//	//// 监听leader变化（使用此方法无法保证强一致性读，仅做leader变化过程观察）
//	//go func() {
//	//	for leader := range myRaft.LeaderCh() {
//	//		if leader {
//	//			atomic.StoreInt64(&global_mata.IsLeader, 1)
//	//		} else {
//	//			atomic.StoreInt64(&global_mata.IsLeader, 0)
//	//		}
//	//	}
//	//}()
//
//	//// todo http服务server
//	//server :=  fasthttp.NewFastHTTPSr(myRaft,fm)
//	//go log.Fatal(fasthttp2.ListenAndServe(httpAddr, server.Router.Handler))
//	//quitCh := make(chan os.Signal, 1)
//	//signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
//	////启动服务
//	////go app.Run()
//	//
//	//<-quitCh
//}