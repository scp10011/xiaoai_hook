package main

import (
	"bytes"
	"flag"
	"fmt"
	goahocorasick "github.com/anknown/ahocorasick"
	"github.com/sevlyar/go-daemon"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

const resFile = "/tmp/mipns/mibrain/mibrain_txt_RESULT_NLP.log"

var controlMethod = map[string]bool{"ch": true, "prev": true, "next": true, "play": true, "pause": true, "toggle": true, "resume": true}

var (
	signal = flag.String("s", "", `向守护程序发送信号:
  stop — 终止进程`)
)

var logFile = flag.String("log", "/dev/null", "启用log默认不输出")
var rpcPort = flag.Int("port", 18888, "jsonrpc端口")
var serverUrl = flag.String("url", "", "服务器URL")
var refreshTime = flag.Int("interval", 60, "刷新间隔时间,0为禁用")
var keys = flag.String("key", "", "拦截关键词用|分割")
var token = flag.String("token", "", "rpc认证token")
var ac = new(goahocorasick.Machine)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

var keyWords [][]rune

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func termHandler(sig os.Signal) error {
	log.Println("正在终止...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	worderArgs := append([]string{"[小爱语音响应拦截器]"}, os.Args[1:]...)
	context := &daemon.Context{
		PidFileName: "/tmp/xiaoai_hook.pid",
		PidFilePerm: 0644,
		LogFileName: *logFile,
		LogFilePerm: 0640,
		WorkDir:     "/tmp/",
		Umask:       027,
		Args:        worderArgs,
	}
	if len(daemon.ActiveFlags()) > 0 {
		d, err := context.Search()
		if err != nil {
			log.Fatalf("无法发送信号到守护程序: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}
	d, err := context.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}
	defer context.Release()
	log.Println("守护程序启动")
	go worker()
	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	log.Println("守护进程已终止")
}

func worker() {
	flag.Parse()
	eventUrl := fmt.Sprintf("%s/xiaoai_hook/event", *serverUrl)
	keyWordUrl := fmt.Sprintf("%s/xiaoai_hook/keyword", *serverUrl)
	if *serverUrl == "" {
		log.Printf("缺少关键参数")
		os.Exit(-1)
	} else {
		log.Printf("上游服务器: %s", *serverUrl)
	}
	log.Printf("正在启动...")
	key := *keys
	if key != "" {
		log.Printf("拦截词列表: %s", key)
		for _, key := range strings.Split(key, "|") {
			keyWords = append(keyWords, bytes.Runes([]byte(key)))
		}
		if err := ac.Build(keyWords); err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}
	}
	if *refreshTime != 0 {
		log.Printf("启动拦截词刷新器,频率: %d秒", *refreshTime)
		go refresh(keyWordUrl)
	}
	log.Printf("启动jsonrpc服务器,监听端口: %d", *rpcPort)
	if *token != "" {
		log.Printf("检测到token开启服务器认证")
	}
	go jsonRpcServer()
	log.Printf("启动日志监控器")
	for {
		monitoring(eventUrl)
	}

}
