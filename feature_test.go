package main

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/zbh255/video-api/model"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestVideoFutures(t *testing.T) {
	model.InitOrm("./testdata/test.db")
	eng := gin.Default()
	InitRouter(eng)
	w := httptest.NewRecorder()
	signBody := `{"user_name":"Tony","user_password":"heap66"}`
	req,err := http.NewRequest("POST","/user/sign",strings.NewReader(signBody))
	if err != nil {
		t.Fatal(err)
	}
	eng.ServeHTTP(w,req)
	assert.Equal(t, Ok.String(),w.Body.String())
	loginBody := signBody
	w = httptest.NewRecorder()
	req,err = http.NewRequest("POST","/user/login",strings.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	eng.ServeHTTP(w,req)
	type Token struct {
		Token string `json:"token"`
	}
	var token Token
	err = json.Unmarshal(w.Body.Bytes(), &token)
	if err != nil {
		t.Fatal(err)
	}
	go eng.Run("localhost:1234")
	dialer := websocket.Dialer{}
	wsUrl := url.URL{
		Scheme: "ws",
		Host:   "localhost:1234",
		Path:   "/video",
	}
	if err != nil {
		t.Fatal(err)
	}
	req,err = http.NewRequest("GET","/video",nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization","Bearer "+token.Token)
	wsConn,_,err := dialer.Dial(wsUrl.String(),req.Header)
	if err != nil {
		t.Fatal(err)
	}
	defer wsConn.Close()
	reqMsg := RequestFrame{
		Type:         "call",
		VideoUrl:     "http://clips.vorwaerts-gmbh.de/big_buck_bunny.mp4",
		VideoOutType: "mp4",
		VideoStart:   "00:00:30",
		VideoEnd:     "00:00:40",
	}
	bytes, err := json.Marshal(reqMsg)
	if err != nil {
		t.Fatal(err)
	}
	err = wsConn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		t.Fatal(err)
	}
	_,buf,err := wsConn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	repMsg := ResponseFrame{}
	err = json.Unmarshal(buf, &repMsg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(repMsg)
	for {
		reqMsg = RequestFrame{
			Type:         "check_status",
			TaskId:       repMsg.TaskId,
		}
		bytes,err = json.Marshal(&reqMsg)
		if err != nil {
			t.Fatal(err)
		}
		err = wsConn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			t.Fatal(err)
		}
		_,buf,err = wsConn.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		var msg ServerErr
		err = json.Unmarshal(buf, &msg)
		if err != nil {
			t.Fatal(err)
		}
		if msg.Code != Ok.Code {
			continue
		} else {
			w = httptest.NewRecorder()
			req,err = http.NewRequest("GET","/video/" + repMsg.SplitVideoId,nil)
			if err != nil {
				t.Fatal(err)
			}
			eng.ServeHTTP(w,req)
		}
	}
}
