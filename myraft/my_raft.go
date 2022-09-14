package myraft

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"tc-kv-cache/fsm"
	"tc-kv-cache/global_mata"
	"time"



	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func NewMyRaft(raftAddr, raftId, raftDir string) (*raft.Raft, *fsm.Fsm, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(raftId)
	// config.HeartbeatTimeout = 1000 * time.Millisecond
	// config.ElectionTimeout = 1000 * time.Millisecond
	// config.CommitTimeout = 1000 * time.Millisecond

	addr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return nil, nil, err
	}
	transport, err := raft.NewTCPTransport(raftAddr, addr, 2, 5*time.Second, os.Stderr)
	if err != nil {
		return nil, nil, err
	}
	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		return nil, nil, err
	}

	// db层
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.db"))
	if err != nil {
		return nil, nil, err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-stable.db"))
	if err != nil {
		return nil, nil, err
	}

	// fsm状态机
	fm := fsm.NewFsm()
	// 初始化raft服务
	rf, err := raft.NewRaft(config, fm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, nil, err
	}

	return rf, fm, nil
}

func Bootstrap(rf *raft.Raft, raftId, raftAddr, raftCluster string) {
	peerArray := strings.Split(raftCluster, ",")
	if len(peerArray) == 0 {
		return
	}
	servers := rf.GetConfiguration().Configuration().Servers
	log.Printf("打印servers：%+v",servers)
	if len(servers) > 0 {
		for _, peerInfo := range peerArray {
			peer := strings.Split(peerInfo, "/")
			addr := peer[1]
			// 填充一下global都addrmap
			//global_mata.AddrMap[addr] = addr[:len(addr)-2] +addr[len(addr)-4:len(addr)-2]   测试用例
			global_mata.AddrMap[addr] = addr[:len(addr)-4]+":8080"
		}
		log.Printf("%+v",global_mata.AddrMap)
		return
	}


	var configuration raft.Configuration
	for _, peerInfo := range peerArray {
		peer := strings.Split(peerInfo, "/")
		id := peer[0]
		addr := peer[1]
		// 填充一下global都addrmap

			//global_mata.AddrMap[addr] = addr[:len(addr)-2] +addr[len(addr)-4:len(addr)-2]
		global_mata.AddrMap[addr] = addr[:len(addr)-4]+"8080"

		server := raft.Server{
			ID:      raft.ServerID(id),
			Address: raft.ServerAddress(addr),
		}
		configuration.Servers = append(configuration.Servers, server)
	}
	log.Printf("addr的map：%+v",global_mata.AddrMap)

	rf.BootstrapCluster(configuration)
	return
}

func GetLeaderIp(rf *raft.Raft ) string {
	addr,_ := rf.LeaderWithID()
	//return strings.Split(string(addr),":")[0]
	return global_mata.AddrMap[string(addr)]

}