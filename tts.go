package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

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
	argv := fmt.Sprintf("{\"text\":\"%s\",\"save\":0}", tts)
	cmd := exec.Command("ubus", "call", "mibrain", "text_to_speech", argv)
	cmd.Start()
	err := cmd.Wait()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
		return false
	}
	return true
}

func waitPlayerTTS(tts string) int {
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
	for i := 1; i <= 100; i++ {
		if playingControl("resume") == 0 {
			break
		}
	}
	log.Printf("拦截默认响应")
}
