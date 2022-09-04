package myraft

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"tc-kv-cache/fsm"
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
	servers := rf.GetConfiguration().Configuration().Servers
	if len(servers) > 0 {
		return
	}
	peerArray := strings.Split(raftCluster, ",")
	if len(peerArray) == 0 {
		return
	}

	var configuration raft.Configuration
	for _, peerInfo := range peerArray {
		peer := strings.Split(peerInfo, "/")
		id := peer[0]
		addr := peer[1]
		server := raft.Server{
			ID:      raft.ServerID(id),
			Address: raft.ServerAddress(addr),
		}
		configuration.Servers = append(configuration.Servers, server)
	}
	rf.BootstrapCluster(configuration)
	return
}
