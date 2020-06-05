package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/url"
	"time"
)

func refresh(url string) {
	second := time.Duration(*refreshTime) * time.Second
	oldText := ""
	for {
		resp, err := netClient.Get(url)
		if err != nil {
			log.Print(err)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			log.Print(err)
			continue
		}
		newText := string(body)
		if oldText != newText {
			var newKeyWords [][]rune
			for _, key := range bytes.Split(body, []byte("|")) {
				newKeyWords = append(newKeyWords, bytes.Runes(key))
			}
			log.Printf("更新拦截词: %s", newText)
			if err := ac.Build(newKeyWords); err != nil {
				log.Fatal(err)
			}
			oldText = newText
		}
		time.Sleep(second)
	}

}

func forwardMsg(upstream string, res []string, answer []string) {
	log.Printf("转发请求...")
	resp, err := netClient.PostForm(upstream, url.Values{"asr": []string{"{}"}, "res": res, "answer": answer})
	if err != nil {
		log.Printf("转发失败: %s\n", err)
	}
	defer func() { _ = resp.Body.Close() }()
	log.Printf("转发成功")
}
