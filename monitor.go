package main

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"log"
)

func monitoring(upstream string) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watch.Close()
	err = watch.Add(resFile)
	if err != nil {
		log.Fatal(err)
	}
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
					for _, t := range ac.MultiPatternSearch(bytes.Runes(b), false) {
						log.Printf("触发拦截词: %s", string(t.Word))
						flag = true
						break
					}
					if flag {
						go waitResumePlayer()
						if resp, err = ioutil.ReadFile(answerFile); err != nil {
							log.Print(err)
							resp = []byte("{\"text\": \"哎呀，小爱刚刚走神啦，请再说一遍吧\"}")
						}
						answer := fastjson.GetString(resp, "text")
						go forwardMsg(upstream, []string{string(b)}, []string{answer})
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
