package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

var mutex sync.Mutex

type ubusResult struct {
	Code int    `json:"code"`
	Info string `json:"info"`
}

type ubusPlayStatus struct {
	Status    int `json:"status"`
	Volume    int `json:"volume"`
	LoopType  int `json:"loop_type"`
	MediaType int `json:"media_type"`
}

func playingControl(s string) int {
	argv := fmt.Sprintf("{\"action\": \"%s\"}", s)
	cmd := exec.Command("ubus", "call", "mediaplayer", "player_play_operation", argv)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	result := &ubusResult{}
	err = json.Unmarshal(output, result)
	if err != nil {
		log.Fatal(err)
	}
	return result.Code
}

func getPlayerStatus() *ubusPlayStatus {
	cmd := exec.Command("ubus", "-t", "1 ", "call", "mediaplayer", "player_get_play_status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	result := &ubusResult{}
	status := &ubusPlayStatus{}
	err = json.Unmarshal(output, result)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal([]byte(result.Info), status)
	if err != nil {
		log.Fatal(err)
	}
	return status
}

func playerTTS(tts string) bool {
	log.Printf("TTS播报: %s", tts)
	argv := fmt.Sprintf("{\"text\":\"%s\",\"save\":0}", tts)
	cmd := exec.Command("ubus", "call", "mibrain", "text_to_speech", argv)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	result := &ubusResult{}
	err = json.Unmarshal(output, result)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("TTS状态: %d, 存储到 %s", result.Code, result.Info)
	return result.Code == 0
}

func waitPlayerTTS(tts string) int {
	mutex.Lock()
	defer func() { mutex.Unlock() }()
	if playerTTS(tts) == false {
		return -1
	}
	temp := getPlayerStatus()
	for temp.MediaType == 1 {
		temp = getPlayerStatus()
		time.Sleep(1 * time.Second)
	}
	return 0
}

func waitResumePlayer() {
	mutex.Lock()
	defer func() { mutex.Unlock() }()
	for i := 1; i <= 100; i++ {
		if playingControl("resume") == 0 {
			break
		}
	}
	log.Printf("拦截响应成功")
}
