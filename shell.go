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

type playSongDetail struct {
	CpOrigin string `json:"cp_origin"`
	CpId     string `json:"cp_id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Duration int    `json:"duration"`
	Position int    `json:"position"`
}

type ubusPlayDetailStatus struct {
	ubusPlayStatus
	Detail playSongDetail   `json:"play_song_detail"`
	Track  []playSongDetail `json:"extra_track_list"`
}

func playingControl(s string) int {
	argv := fmt.Sprintf("{\"action\": \"%s\"}", s)
	cmd := exec.Command("ubus", "call", "mediaplayer", "player_play_operation", argv)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Print(err)
		return -1
	}
	result := &ubusResult{}
	err = json.Unmarshal(output, result)
	if err != nil {
		log.Print(err)
		return -1
	}
	return result.Code
}

func getPlayerStatus() *ubusPlayStatus {
	cmd := exec.Command("ubus", "-t", "1 ", "call", "mediaplayer", "player_get_play_status")
	output, err := cmd.CombinedOutput()
	result := &ubusResult{}
	status := &ubusPlayStatus{}
	if err != nil {
		log.Print(err)
		return status
	}
	err = json.Unmarshal(output, result)
	if err != nil {
		log.Print(err)
	}
	err = json.Unmarshal([]byte(result.Info), status)
	if err != nil {
		log.Print(err)
	}
	return status
}

func getPlayerDetailStatus() *ubusPlayDetailStatus {
	cmd := exec.Command("ubus", "-t", "1 ", "call", "mediaplayer", "player_get_play_status")
	output, err := cmd.CombinedOutput()
	result := &ubusResult{}
	status := &ubusPlayDetailStatus{}
	if err != nil {
		log.Print(err)
		return status
	}
	err = json.Unmarshal(output, result)
	if err != nil {
		log.Print(err)
	}
	err = json.Unmarshal([]byte(result.Info), status)
	if err != nil {
		log.Print(err)
	}
	trackLength := len(status.Track)
	for i := 0; i < trackLength; {
		if status.Track[i].CpOrigin == "" {
			trackLength--
			status.Track[i] = status.Track[trackLength]
			status.Track = status.Track[:trackLength]
		} else {
			i++
		}
	}
	return status
}

func playerTTS(tts string) bool {
	log.Printf("TTS播报: %s", tts)
	argv := fmt.Sprintf("{\"text\":\"%s\",\"save\":0}", tts)
	cmd := exec.Command("ubus", "call", "mibrain", "text_to_speech", argv, "&")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Print(err)
		return false
	}
	//result := &ubusResult{}
	//err = json.Unmarshal(output, result)
	//if err != nil {
	//	log.Print(err)
	//	return false
	//}
	// log.Printf("TTS状态: %d, URL: %s", result.Code, result.Info)
	return true
}

func editVolume(v float64) bool {
	if 0 <= v && v <= 100 {
		cmd := exec.Command("mphelper", "volume_set", fmt.Sprint(v))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Print(err)
			return false
		}
		result := &Result{}
		err = json.Unmarshal(output, result)
		if err != nil {
			log.Print(err)
			return false
		}
		log.Printf("设置音量到: %v", v)
		return result.Code == 0
	} else {
		return false
	}
}

func adjustVolume(mode string) bool {
	var operator string
	if mode == "up" {
		operator = "volume_up"
		log.Printf("增大音量")
	} else {
		operator = "volume_down"
		log.Printf("降低音量")
	}
	cmd := exec.Command("mphelper", operator)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Print(err)
		return false
	}
	result := &Result{}
	err = json.Unmarshal(output, result)
	if err != nil {
		log.Print(err)
		return false
	}
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
		time.Sleep(50 * time.Microsecond)
	}
	return 0
}

func waitResumePlayer() {
	log.Printf("尝试拦截默认响应...")
	mutex.Lock()
	defer func() { mutex.Unlock() }()
	temp := getPlayerStatus()
	for temp.Status == 1 {
		_ = playingControl("resume")
	}
	_ = playerTTS("稍等") // 尝试顶掉卡住的输出
	log.Printf("拦截响应成功")
}
