package main

import (
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func refresh(url string) {
	second := time.Duration(*refreshTime) * time.Second
	for {
		time.Sleep(second)
		oldkeyWords := strings.Join(keyWords, "|")
		resp, err := netClient.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		text := string(body)
		if oldkeyWords != text {
			log.Printf("更新拦截词: %s", text)
		}
		keyWords = strings.Split(text, "|")
	}

}
