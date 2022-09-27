package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	path2 "path"
	"tc-kv-cache/iface"
	"time"
)
func TellLeader(host,path string,param interface{})  error{
	var err error
	switch path {
	case "add":
		p := param.(iface.KV)
		err = TellLeaderAdd(host,path,p.Key,p.Value)
		if err != nil {
			return err
		}
	case "del":
		key := param.(string)
		err = TellLeaderDel(host,path,key)
		if err != nil {
			return err
		}
	case "zadd":
		sv:= param.(iface.TellBody).Val.(iface.SV)
		key := param.(iface.TellBody).Key
		err = TellLeaderZadd(host,path,key,sv)
		if err != nil {
			return err
		}
	case "batch":
		kvs := param.([]iface.KV)
		err = TellLeaderBatch(host,path,kvs)
		if err != nil {
			return err
		}
	case "zrmv":
		key := param.(string)
		err = TellLeaderZrem(host,path,key)
		if err != nil {
			return err
		}

	}
	return nil
}
func TellLeaderAdd(host,path string,ky,val string)  error{
	apiUrl := url.URL{
		Host:   host,
		Path:   path,
		Scheme: "http",
	}
	m := make(map[string]string)
	m["key"] = ky
	m["value"] = val
	jsonstr,_ := json.Marshal(m)
	fmt.Println("url是：",apiUrl.String())
	resp,err := httpClient.Post(apiUrl.String(),"Content-Type: application/json",bytes.NewReader(jsonstr))
	if err != nil || resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("err: %v, resp : %+v ",err,resp))

	}
	log.Println(resp)
	return nil
}

func TellLeaderDel(host,path ,key string) error {
	apiUrl := url.URL{
		Scheme:      "http",
		Host:        host,
		Path:        path,
	}
	apiUrl.Path = path2.Join(apiUrl.Path,key)


	fmt.Println("url是：",apiUrl.String())
	resp,err := httpClient.Get(apiUrl.String())
	if err != nil || resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("err: %v, resp : %+v ",err,resp))

	}
	log.Println(resp)
	return nil

}
// zadd
func TellLeaderZadd(host,path,key string,sv iface.SV)  error{
	apiUrl := url.URL{
		Host:   host,
		Path:   path,
		Scheme: "http",
	}
	apiUrl.Path = path2.Join(apiUrl.Path,key)

	jsonstr,_ := json.Marshal(sv)
	fmt.Println("url是：",apiUrl.String())
	resp,err := httpClient.Post(apiUrl.String(),"Content-Type: application/json",bytes.NewReader(jsonstr))
	if err != nil || resp.StatusCode != 200 {

		return errors.New(fmt.Sprintf("err: %v, resp : %+v ",err,resp))
	}
	log.Println(resp)
	return nil
}
// batch
func TellLeaderBatch(host,path string,kvs []iface.KV)  error{
	apiUrl := url.URL{
		Host:   host,
		Path:   path,
		Scheme: "http",
	}

	jsonstr,_ := json.Marshal(kvs)
	fmt.Println("url是：",apiUrl.String())
	resp,err := httpClient.Post(apiUrl.String(),"Content-Type: application/json",bytes.NewReader(jsonstr))
	if err != nil || resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("err: %v, resp : %+v ",err,resp))

	}
	log.Println(resp)
	return nil
}
// zrem
func TellLeaderZrem(host,path ,key string) error {
	apiUrl := url.URL{
		Scheme:      "http",
		Host:        host,
		Path:        path,
	}
	apiUrl.Path = path2.Join(apiUrl.Path,key)


	fmt.Println("url是：",apiUrl.String())
	resp,err := httpClient.Get(apiUrl.String())
	if err != nil || resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("err: %v, resp : %+v ",err,resp))

	}
	log.Println(resp)
	return nil

}

var (
	httpClient *http.Client
)
func init()  {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     5 * time.Second,
		},
	}
httpClient = client
}
