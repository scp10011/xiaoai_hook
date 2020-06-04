package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"log"
	"net/url"
)

func monitoring(upstream string) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watch.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = watch.Add(resFile)
	var (
		flag         bool
		resp         []byte
		oldRequestId string
	)
	for {
		select {
		case ev := <-watch.Events:
			{
				if ev.Op&fsnotify.Write == fsnotify.Write {
					b, err := ioutil.ReadFile(resFile)
					if err != nil {
						log.Print(err)
						continue
					}
					requestId := fastjson.GetString(b, "meta", "request_id")
					if oldRequestId == requestId {
						continue
					}
					oldRequestId = requestId
					log.Printf("检测到文件变动")
					flag = false
					text := fastjson.GetString(b, "response", "answer", "0", "text")
					if len(text) == 0 {
						continue
					}
					log.Printf("检测语句: %s", text)
					for _, t := range ac.MultiPatternSearch([]rune(text), false) {
						log.Printf("触发拦截词: %s", string(t.Word))
						flag = true
					}
					if flag {
						if resp, err = ioutil.ReadFile(answerFile); err != nil {
							log.Print(err)
							resp = []byte("{\"text\": \"哎呀，小爱刚刚走神啦，请再说一遍吧\"}")
						}
						answer := fastjson.GetString(resp, "text")
						log.Printf("尝试拦截默认响应...")
						go forwardMsg(upstream, []string{string(b)}, []string{answer})
						waitResumePlayer()
					}
				} else {
					return
				}
			}
		case <-stop:
			return
		case err := <-watch.Errors:
			{
				log.Println("error : ", err)
				return
			}
		}
	}
}

func forwardMsg(upstream string, res []string, answer []string) {
	log.Printf("转发请求...")
	resp, err := netClient.PostForm(upstream, url.Values{"asr": []string{"{}"}, "res": res, "answer": answer})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Printf("转发成功")
}
