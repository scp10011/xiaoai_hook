package main

import (
	"context"
	"fmt"
	"github.com/intel-go/fastjson"
	"github.com/osamingo/jsonrpc"
	"log"
	. "net/http"
)

type (
	ttsBroadcast       struct{}
	ttsBroadcastParams struct {
		Token string `json:"token"`
		Msg   string `json:"msg"`
	}
	Result struct {
		Code int `json:"code"`
	}
	playerStatus       struct{}
	playerStatusParams struct {
		Token string `json:"token"`
	}
	playerStatusResult struct {
		Code int `json:"code"`
		ubusPlayDetailStatus
	}
	playerControl       struct{}
	playerControlParams struct {
		Token  string `json:"token"`
		Method string `json:"method"`
	}
	volumeControl       struct{}
	volumeControlParams struct {
		Token    string      `json:"token"`
		Operator interface{} `json:"operator"`
	}
)

func authenticate(t string) bool {
	if *token != "" && *token == t {
		return true
	} else {
		return false
	}

}

func (h ttsBroadcast) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p ttsBroadcastParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if !authenticate(p.Token) {
		return Result{Code: -5}, nil
	}
	original := getPlayerStatus()
	if original.Status == 1 {
		playingControl("pause")
	}
	code := waitPlayerTTS(p.Msg)
	if original.Status == 1 {
		playingControl("play")
	}
	return Result{
		Code: code,
	}, nil
}

func (h playerStatus) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p playerStatusParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if !authenticate(p.Token) {
		return playerStatusResult{Code: -5}, nil
	}
	status := getPlayerDetailStatus()
	return playerStatusResult{
		0, *status,
	}, nil
}

func (h playerControl) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p playerControlParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if !authenticate(p.Token) {
		return Result{Code: -1}, nil
	}
	if controlMethod[p.Method] {
		return Result{Code: playingControl(p.Method)}, nil
	} else {
		return Result{Code: -1}, nil
	}
}

func (h volumeControl) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p volumeControlParams
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if !authenticate(p.Token) {
		return Result{Code: -1}, nil
	}
	flag := false
	switch p.Operator.(type) {
	case string:
		flag = adjustVolume(p.Operator.(string))
	case float64:
		flag = editVolume(p.Operator.(float64))
	}
	if flag {
		return Result{Code: 0}, nil
	} else {
		return Result{Code: -1}, nil
	}
}

func jsonRpcServer() {
	mr := jsonrpc.NewMethodRepository()
	if err := mr.RegisterMethod("TTS", ttsBroadcast{}, ttsBroadcastParams{}, Result{}); err != nil {
		log.Fatalln(err)
	}
	if err := mr.RegisterMethod("STATUS", playerStatus{}, playerStatusParams{}, playerStatusResult{}); err != nil {
		log.Fatalln(err)
	}
	if err := mr.RegisterMethod("CONTROL", playerControl{}, playerControlParams{}, Result{}); err != nil {
		log.Fatalln(err)
	}
	if err := mr.RegisterMethod("VOLUME", volumeControl{}, volumeControlParams{}, Result{}); err != nil {
		log.Fatalln(err)
	}
	Handle("/", mr)
	port := fmt.Sprintf(":%d", *rpcPort)
	if err := ListenAndServe(port, DefaultServeMux); err != nil {
		log.Fatalln(err)
	}
}
