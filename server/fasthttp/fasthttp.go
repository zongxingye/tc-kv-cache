package fasthttp

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/spf13/cast"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"tc-kv-cache/client"
	"tc-kv-cache/myraft"

	//e "tc-kv-cache/db/leveldb"
	"tc-kv-cache/fsm"
	"tc-kv-cache/global_mata"
	"tc-kv-cache/iface"
	"time"

	//e "wecc/engines/rosedb"
	//e "wecc/engines/simple"


	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var (
	addr     = flag.String("addr", ":8080", "TCP address to listen to")
	compress = flag.Bool("compress", false, "Whether to enable transparent response compression")
)
type HttpServer struct {
	ctx *raft.Raft
	fsm    *fsm.Fsm
	Router *router.Router
}

func NewFastHTTPSr(ctx *raft.Raft,	fsm *fsm.Fsm)HttpServer {
	server := HttpServer{
		ctx: ctx,
		fsm: fsm,
	}
	flag.Parse()

	r := router.New()
	server.Register(r)
	server.Router = r
	return  server
}

func (h *HttpServer)Register(r *router.Router) {
	//var engine = e.NewEngine()
	//engine :=h.fsm.DataBase.Engine
	r.POST("/updateCluster", func(ctx *fasthttp.RequestCtx) {
		var req iface.UpdateCluster
		body := ctx.PostBody()
		if e := json.Unmarshal(body, &req); e != nil {
			ctx.Error(e.Error(), fasthttp.StatusBadRequest)
			return
		}
		// 初始化initraft
		raftId:= req.Index
		raftDir := "node/raft_" + cast.ToString(raftId)
		os.MkdirAll(raftDir, 0700)
		raftIp := req.Hosts[req.Index-1]
		// 初始化raft
		myRaft, fm, err := myraft.NewMyRaft(raftIp+":8088", cast.ToString(raftId), raftDir)
		//myRaft, fm, err := myraft.NewMyRaft("0.0.0.0:8088", cast.ToString(raftId), raftDir)
		if err != nil {
			fmt.Println("NewMyRaft error ", err)
			os.Exit(1)
			return
		}
		log.Println("NewMyRaft:",raftIp+":8088")
		h.fsm = fm
		h.ctx = myRaft
		clusterGroup := make([]string,0)
		for i,v := range req.Hosts{
			clusterGroup = append(clusterGroup, cast.ToString(i+1) + "/"+v+":8088")
		}
		myraft.Bootstrap(h.ctx, "todo", "todo",strings.Join(clusterGroup,",") )

		// 监听leader变化（使用此方法无法保证强一致性读，仅做leader变化过程观察）
		go func() {
			for leader := range myRaft.LeaderCh() {
				if leader {
					atomic.StoreInt64(&global_mata.IsLeader, 1)
				} else {
					atomic.StoreInt64(&global_mata.IsLeader, 0)
				}
			}
		}()
	})

	r.GET("/init", func(ctx *fasthttp.RequestCtx) {
		engine :=h.fsm.DataBase.Engine
		engine.Init(context.Background())
		ctx.WriteString("ok")
	})

	r.GET("/query/{key}", func(ctx *fasthttp.RequestCtx) {

		key, ok := ctx.UserValue("key").(string)
		if !ok {
			ctx.Error("未获取到key", fasthttp.StatusNotImplemented)
			return
		}
		engine :=h.fsm.DataBase.Engine
		value, ok, err := engine.Get(context.Background(), key)
		if err != nil {
			log.Println("engine get err: ", err)
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError), fasthttp.StatusInternalServerError)
			return
		}
		if ok {
			//成功返回结果
			ctx.WriteString(value)
			return
		} else {
			//不存在返回404
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound), fasthttp.StatusNotFound)
			return
		}
	})
	r.POST("/add", func(ctx *fasthttp.RequestCtx) {

		// parse JSON body
		var add = iface.KV{}
		body := ctx.PostBody()
		if e := json.Unmarshal(body, &add); e != nil {
			ctx.Error(e.Error(), fasthttp.StatusBadRequest)
			return
		}

		// 不是leader应该直接通知给leader写入
		if atomic.LoadInt64(&global_mata.IsLeader) == 0 {
			//ctx.WriteString("not leader")
			leaderHost:=myraft.GetLeaderIp(h.ctx)
			err := client.TellLeader(leaderHost,"add",add)
			if err != nil {
				ctx.WriteString("error:"+err.Error())
			}else {
				ctx.WriteString("ok")
			}
			return
		}
		//err := engine.Add(context.Background(), add.Key, add.Value)
		//if err != nil {
		//	ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		//	return
		//}
		// raft.apply
		data := "set"+"@"+add.Key+"@"+add.Value
		future := h.ctx.Apply([]byte(data),5*time.Second)
		if err := future.Error(); err != nil {
			ctx.WriteString("error:"+err.Error())
			return
		}
		ctx.WriteString("ok")
	})

	//只定义了200
	r.GET("/del/{key}", func(ctx *fasthttp.RequestCtx) {

		key, ok := ctx.UserValue("key").(string)
		if !ok {
			ctx.Error("未获取到key", fasthttp.StatusBadGateway)
			return
		}
		// leader转发
		if atomic.LoadInt64(&global_mata.IsLeader) == 0 {
			leaderHost:=myraft.GetLeaderIp(h.ctx)
			err:= client.TellLeader(leaderHost,"del",key)
			if err != nil {
				ctx.WriteString("error:"+err.Error())
			}else {
				ctx.WriteString("ok")
			}
			return
		}
		data := "del"+"@"+key
		future := h.ctx.Apply([]byte(data),5*time.Second)
		if err := future.Error(); err != nil {
			ctx.WriteString("error:"+err.Error())
			return
		}
		ctx.WriteString("ok")
		return
	})

	// 404 The key can not be found in the cache
	r.POST("/list", func(ctx *fasthttp.RequestCtx) {
		// parse JSON body
		var req []string
		body := ctx.PostBody()
		if e := json.Unmarshal(body, &req); e != nil {
			ctx.Error(e.Error(), 500)
			return
		}
		engine :=h.fsm.DataBase.Engine
		res, err := engine.List(context.Background(), req)
		if err != nil {
			ctx.Error(err.Error(), 500)
			return
		}
		if len(res) == 0 {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound), fasthttp.StatusNotFound)
			return
		}
		resbyte, err := json.Marshal(res)
		if err != nil {
			ctx.Error(err.Error(), 500)
			return
		}
		ctx.Write(resbyte)
	})

	//400 Insert key and value failed
	r.POST("/batch", func(ctx *fasthttp.RequestCtx) {
		// parse JSON body
		var req []iface.KV
		body := ctx.PostBody()
		if e := json.Unmarshal(body, &req); e != nil {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)

			return
		}

		// leader转发
		if atomic.LoadInt64(&global_mata.IsLeader) == 0 {
			leaderHost:=myraft.GetLeaderIp(h.ctx)
			err:= client.TellLeader(leaderHost,"batch",req)
			if err != nil {
				ctx.WriteString("error:"+err.Error())
			}else {
				ctx.WriteString("ok")
			}
			return
		}
		reqByt,_ := json.Marshal(req)
		data := "batch"+"@"+string(reqByt)
		future := h.ctx.Apply([]byte(data),5*time.Second)
		if err := future.Error(); err != nil {
			ctx.WriteString("error:"+err.Error())
			return
		}
		ctx.WriteString("ok")
		//engine :=h.fsm.DataBase.Engine
		//err := engine.Batch(context.Background(), req)
		//if err != nil {
		//	ctx.Error(fasthttp.StatusMessage(fasthttp.StatusBadRequest), fasthttp.StatusBadRequest)
		//
		//	return
		//}
		//ctx.WriteString("ok")
	})

	r.POST("/zadd/{key}", func(ctx *fasthttp.RequestCtx) {
		// parse JSON body
		var sv iface.SV
		key, ok := ctx.UserValue("key").(string)
		if !ok {
			ctx.Error("未获取到key", 400)
			return
		}
		body := ctx.PostBody()
		if e := json.Unmarshal(body, &sv); e != nil {
			ctx.Error(e.Error(), 500)
			return
		}

		// leader转发
		tell := iface.TellBody{
			Key: key,
			Val: sv,
		}
		if atomic.LoadInt64(&global_mata.IsLeader) == 0 {
			leaderHost:=myraft.GetLeaderIp(h.ctx)
			err:= client.TellLeader(leaderHost,"zadd",tell)
			if err != nil {
				ctx.WriteString("error:"+err.Error())
			}else {
				ctx.WriteString("ok")
			}
			return
		}
		svByt,_ := json.Marshal(sv)
		data := "zadd"+"@"+key +"@"+string(svByt)
		future := h.ctx.Apply([]byte(data),5*time.Second)
		if err := future.Error(); err != nil {
			ctx.WriteString("error:"+err.Error())
			return
		}
		ctx.WriteString("ok")
		//engine :=h.fsm.DataBase.Engine
		//err := engine.ZAdd(context.Background(), key, sv)
		//if err != nil {
		//	ctx.Error(err.Error(), 500)
		//	return
		//}
		//ctx.WriteString("ok")
	})

	//404 he key can not be found in the cache
	r.POST("/zrange/{key}", func(ctx *fasthttp.RequestCtx) {
		// parse JSON body
		var sr iface.ScRange
		key, ok := ctx.UserValue("key").(string)
		if !ok {
			ctx.Error("未获取到key", 502)
			return
		}
		body := ctx.PostBody()
		if e := json.Unmarshal(body, &sr); e != nil {
			ctx.Error(e.Error(), 500)
			return
		}
		engine :=h.fsm.DataBase.Engine
		sv, err := engine.ZRange(context.Background(), key, sr)
		if err != nil {
			ctx.Error(err.Error(), 500)
			return
		}
		if len(sv) == 0 {
			ctx.Error("not found", 404)
			return
		}
		resbyte, err := json.Marshal(sv)
		if err != nil {
			ctx.Error(err.Error(), 500)
			return
		}
		ctx.Write(resbyte)
	})

	r.GET("/zrmv/{key}/{value}", func(ctx *fasthttp.RequestCtx) {
		// parse JSON body
		key, ok := ctx.UserValue("key").(string)
		if !ok {
			ctx.Error("未获取到key", 502)
			return
		}
		value, ok := ctx.UserValue("value").(string)
		if !ok {
			ctx.Error("未获取到value", 502)
			return
		}
		if atomic.LoadInt64(&global_mata.IsLeader) == 0 {
			leaderHost:=myraft.GetLeaderIp(h.ctx)
			err:= client.TellLeader(leaderHost,"zrmv",key+"/"+value)
			if err != nil {
				ctx.WriteString("error:"+err.Error())
			}else {
				ctx.WriteString("ok")
			}
			return
		}


		data := "zrmv"+"@"+key +"@"+value
		future := h.ctx.Apply([]byte(data),5*time.Second)
		if err := future.Error(); err != nil {
			ctx.WriteString("error:"+err.Error())
			return
		}
		ctx.WriteString("ok")
		//engine :=h.fsm.DataBase.Engine
		//err := engine.ZRmv(context.Background(), key, value)
		//if err != nil {
		//	ctx.Error(err.Error(), 500)
		//	return
		//}
		//ctx.WriteString("ok")
	})

	//---------------------test--------------------
	r.GET("/testq", func(ctx *fasthttp.RequestCtx) {
		rand.Seed(time.Now().Unix())
		engine :=h.fsm.DataBase.Engine
		value, ok, err := engine.Get(context.Background(), fmt.Sprintf("batch%d", rand.Intn(30000000)))
		if err != nil {
			log.Println("engine get err: ", err)
		}
		if ok {
			//成功返回结果
			ctx.WriteString(value)
			return
		} else {
			//不存在返回404
			ctx.Error("未获取到key", 404)
			return
		}
	})
}
