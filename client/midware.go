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
	"time"
)

func TellLeaderSet(host,path string,ky,val string)  error{
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
	resp,err := httpClient.Post(apiUrl.String(),"application/x-www-form-urlencoded",bytes.NewReader(jsonstr))
	if err != nil || resp.StatusCode != 200 {
		return errors.New("err")
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
