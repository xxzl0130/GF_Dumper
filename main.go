package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/pkg/errors"
	cipher "github.com/xxzl0130/GF_cipher"
)

type Dumper struct{
	sign string
	port int
}

type Uid struct {
	Sign            string `json:"sign"`
}

func main() {
	dumper := &Dumper{
		sign : "",
		port : 8888,
	}
	if err := dumper.Run(); err != nil {
		fmt.Printf("程序启动失败 -> %+v", err)
	}
}

func (dumper *Dumper) Run() error {
	localhost, err := dumper.getLocalhost()
	if err != nil {
		fmt.Printf("获取代理地址失败 -> %+v", err)
		return err
	}

	fmt.Printf("代理地址 -> %s:%d\n", localhost, dumper.port)

	srv := goproxy.NewProxyHttpServer()
	srv.OnResponse(dumper.condition()).DoFunc(dumper.onResponse)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", dumper.port), srv); err != nil {
		fmt.Printf("启动代理服务器失败 -> %+v\n", err)
	}

	return nil
}

func (dumper *Dumper) onResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	fmt.Printf("处理请求响应 -> %s\n", path(ctx.Req))
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("读取响应数据失败 -> %+v\n", err)
		return resp
	}
	if len(body) == 0 || body[0] != byte(35){
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	}

	if strings.HasSuffix(ctx.Req.URL.Path,"/Index/getDigitalSkyNbUid") || strings.HasSuffix(ctx.Req.URL.Path, "/Index/getUidTianxiaQueue") || strings.HasSuffix(ctx.Req.URL.Path,"/Index/getUidEnMicaQueue"){
		// 处理sign
		dumper.processSign(body)
	}else{
		data, err := cipher.AuthCodeDecodeB64(string(body)[1:], dumper.sign, true)
		if err != nil {
			fmt.Printf("解析数据失败 -> %+v\n", err)
		}else{
			
			_ = ioutil.WriteFile(fmt.Sprintf("response.%s.%d.json", strings.Replace(ctx.Req.URL.Path,"/","-",-1), time.Now().Unix()), []byte(data), 0)
		}
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return resp
}

func (dumper *Dumper) processSign(body []byte){
	data, err := cipher.AuthCodeDecodeB64Default(string(body)[1:])
	if err != nil {
		fmt.Printf("解析Uid数据失败 -> %+v\n", err)
		return
	}
	uid := Uid{}
	if err := json.Unmarshal([]byte(data), &uid); err != nil {
		fmt.Printf("解析JSON数据失败 -> %+v\n", err)
		return
	}
	dumper.sign = uid.Sign
}

func (dumper *Dumper) condition() goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		fmt.Printf("请求 -> %v\n", path(req))
		if strings.HasSuffix(req.Host, "ppgame.com") || strings.HasSuffix(req.Host, "sn-game.txwy.tw")  || strings.HasSuffix(req.Host, "girlfrontline.co.kr") || strings.HasSuffix(req.Host, "sunborngame.com") || strings.HasSuffix(req.Host, "sn-game.txwy.tw") {
			//if strings.HasSuffix(req.URL.Path, "/Index/index") || strings.HasSuffix(req.URL.Path, "/Index/getDigitalSkyNbUid") || strings.HasSuffix(req.URL.Path, "/Index/getUidTianxiaQueue") || strings.HasSuffix(req.URL.Path,"/Index/getUidEnMicaQueue"){
				return true
			//}
		}
		return false
	}
}

func (dumper *Dumper) getLocalhost() (string, error) {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		return "", errors.WithMessage(err, "连接 114.114.114.114 失败")
	}
	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return "", errors.WithMessage(err, "解析本地主机地址失败")
	}
	return host, nil
}

func path(req *http.Request) string {
	if req.URL.Path == "/" {
		return req.Host
	}
	return req.Host + req.URL.Path
}
