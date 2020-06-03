package main

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"net/url"
	"os"
	"strings"
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
	_, err = os.Stat(resFile)
	var fp *os.File
	if os.IsNotExist(err) {
		fp, err = os.Create(resFile)
		defer fp.Close()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	} else if err != nil {
		log.Fatal(err)
	} else {
		fp, err = os.Open(resFile)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer fp.Close()
		_, err = fp.Seek(0, 2)
	}
	resFileScanner := bufio.NewScanner(fp)
	resFileScanner.Split(bufio.ScanLines)
	var (
		success bool
		flag    bool
	)
	err = watch.Add(resFile)
	for {
		select {
		case ev := <-watch.Events:
			{
				if ev.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("检测到文件变动")
					success = resFileScanner.Scan()
					if success == false {
						err = resFileScanner.Err()
						if err == nil {
							return
						} else {
							log.Fatal(err)
						}
					}
					resLine := resFileScanner.Text()
					flag = false
					for _, key := range keyWords {
						if strings.Contains(resLine, key) == true {
							log.Printf("触发拦截词语: %s", key)
							flag = true
							break
						}
					}
					// 兼容其他库
					if flag {
						log.Printf("尝试拦截默认响应...")
						go waitResumePlayer()
						go forwardMsg(upstream, []string{resLine})
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

func forwardMsg(upstream string, res []string) {
	log.Printf("转发请求...")
	resp, err := netClient.PostForm(upstream, url.Values{"asr": []string{"{}"}, "res": res})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Printf("转发成功")
}
