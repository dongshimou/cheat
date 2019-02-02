package protocol

import (
	"encoding/json"
)

type WSLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type EventType int32

const (
	//错误的指令
	UC_ErrCmd EventType = iota

	//注册,登录指令
	UC_SignCmd

	//房间相关指令
	UC_RoomCmd

	//游戏相关指令
	UC_PlayCmd

	//...
)

type WsRequest struct {
	Cmd  EventType
	Data json.RawMessage
}

type WsResponse struct {
	Cmd  EventType
	Data interface{}
}
