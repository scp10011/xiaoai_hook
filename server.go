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
		Msg   string `json:"name"`
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
		ubusPlayStatus
	}
	playerControl       struct{}
	playerControlParams struct {
		Token  string `json:"token"`
		Method string `json:"method"`
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
	status := getPlayerStatus()
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
	Handle("/jsonrpc", mr)
	port := fmt.Sprintf(":%d", *rpcPort)
	if err := ListenAndServe(port, DefaultServeMux); err != nil {
		log.Fatalln(err)
	}
}
