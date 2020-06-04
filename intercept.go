package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"time"
)

func refresh(url string) {
	second := time.Duration(*refreshTime) * time.Second
	oldText := ""
	for {
		resp, err := netClient.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
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
