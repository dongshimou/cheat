package protocol

import (
	"encoding/json"
)

type SignCmd int

const (
	SC_Nil SignCmd = iota

	SC_Login  //登录
	SC_Logout //退出

	SC_Signin  //注册
	SC_Signout //注销

)

type SignRequest struct {
	Type     SignCmd
	Username string
	Password string
	Email    string
}

type ChatCmd int

const (
	CC_Nil ChatCmd = iota

	CC_Private //私聊
	CC_Room    //房间内
	CC_Server  //全服
)

type ChatRequest struct {
	Type ChatCmd

	Context string

	TargetId int64
}

type RoomCmd int

const (
	RC_Nil RoomCmd = iota
	RC_Join
	RC_Create
	RC_Rank
	RC_List
	RC_Exit
)

type RoomRequest struct {
	Type RoomCmd
	//加入的时候需要id
	RoomId int64
}

type RoomResponse struct {
	Type RoomCmd
	Data interface{}
}

type RoomResListSingle struct {
	Id          int64
	Name        string
	PlayerCount int
	NeedPass    bool
}
type RoomResCreate struct {
	Id          int64
	Name        string
	PlayerCount int
	Password    string
}

type ErrorResponse struct {
	Code int
	Msg  string
}

type UserPlayCmd int

const (
	UPC_Error UserPlayCmd = iota
	UPC_Ready
	UPC_Watch
	UPC_Cheat
	UPC_Check
	UPC_Throw
	UPC_Exit
)

type PlayRequest struct {
	PlayCmd UserPlayCmd

	//玩家
	PlayerId int64

	//投注额
	Amount uint64

	//是否已看牌
	Watch bool

	//投注价值
	Volume uint64

	//是否查人
	CheckId int64

	//是否有人输了
	LoserId int64
}

type EventType int32

const (
	//错误
	UC_ErrCmd EventType = iota

	//注册,登录
	UC_SignCmd

	//房间相关
	UC_RoomCmd

	//游戏相关
	UC_PlayCmd

	//通知相关
	UC_NotifyCmd

	//聊天
	UC_ChatCmd
	//...
)

type WsRequest struct {
	Type EventType
	Data json.RawMessage
	Uid int64
}

type WsResponse struct {
	Type EventType
	Data interface{}
}
