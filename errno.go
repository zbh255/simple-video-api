package main

import "github.com/goccy/go-json"

type ServerErr struct {
	Code int
	Msg string
}

func (e *ServerErr) String() string {
	b,_ := json.Marshal(e)
	return string(b)
}

func NewServerErr(code int, msg string) *ServerErr {
	return &ServerErr{
		Code: code,
		Msg:  msg,
	}
}

var (
	Ok = NewServerErr(200, "OK")
	ErrJsonParam = NewServerErr(10010,"json参数不正确")
	ErrServer = NewServerErr(10011,"服务器错误")
	ErrUserSign = NewServerErr(10012,"用户未登录")
	ErrVideoTask = NewServerErr(10021,"视频任务不存在或者已经处理完毕")
	ErrVideoTaskHandle = NewServerErr(10023,"视频任务处理出错")
	ErrVideoTaskNoOk = NewServerErr(10022,"视频任务正在处理")
)