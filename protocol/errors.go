package protocol

import (
	"cheat/util"
	"errors"
)

const (
	SuccessCode = 0
	SuccessStr  = "成功"

	ErrUnDefineErrorCode = 99999

	ErrNotFoundUserCode = 10001
	ErrNotFoundUserStr  = "没有这个用户"

	ErrWrongPasswordCode = 10002
	ErrWrongPasswordStr  = "密码错误"

	ErrRoomNotExistCode = 30001
	ErrRoomNotExistStr  = "房间不存在"

	ErrPlayerUnknowStatusCode = 40001
	ErrPlayerUnknowStatusStr  = "该玩家状态异常"

	ErrPlayerLosedCode = 40002
	ErrPlayerLosedStr  = "该玩家已经输了"

	ErrPlayerExitedCode = 40003
	ErrPlayerExitedStr  = "该玩家已经退出"

	ErrPlayerCmdRejectCode = 40004
	ErrPlayerCmdRejectStr  = "当前不能操作"
)

var errMap = map[int]string{}

func init() {
	errMap[SuccessCode] = SuccessStr
	errMap[ErrNotFoundUserCode] = ErrNotFoundUserStr
	errMap[ErrRoomNotExistCode] = ErrRoomNotExistStr
	errMap[ErrWrongPasswordCode] = ErrWrongPasswordStr
	errMap[ErrPlayerUnknowStatusCode] = ErrPlayerUnknowStatusStr
	errMap[ErrPlayerLosedCode] = ErrPlayerLosedStr
	errMap[ErrPlayerExitedCode] = ErrPlayerExitedStr
	errMap[ErrPlayerCmdRejectCode] = ErrPlayerCmdRejectStr

}

type InnerError struct {
	Code int `json:"code"`
	error
}

func NewErr(code int) *InnerError {
	return NewErrs(code, errMap[code])
}

func NewErrs(code int, args ...interface{}) *InnerError {
	return &InnerError{Code: code, error: errors.New(util.ToString(args...))}
}

func NewError(args ...interface{})*InnerError{
	return NewErrs(ErrUnDefineErrorCode,args...)
}

func (e *InnerError) Warp(code int, args ...interface{}) {
	e.Code = code
	e.error = errors.New(util.ToString(args...))
}
