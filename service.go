package main

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zbh255/video-api/model"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type RequestFrame struct {
	Type         string `json:"type"`
	TaskId       int    `json:"task_id"`
	VideoId      string `json:"video_id"`
	VideoUrl     string `json:"video_url"`
	VideoOutType string `json:"video_out_type"`
	VideoStart   string `json:"video_start"`
	VideoEnd     string `json:"video_end"`
}

type ResponseFrame struct {
	Type         string        `json:"type"`
	TaskId       int           `json:"task_id"`
	StateComment string        `json:"state_comment"`
	VideoSize    int           `json:"video_size"`
	SplitTime    time.Duration `json:"split_time"`
	SplitVideoId string        `json:"split_video_id"`
}

type VideoImpl struct {
	mu sync.RWMutex
	pool *TaskPool
	taskIdStart int
	taskCollection map[int]VideoSplitTask
}

type VideoSplitTask struct {
	SplitVideoFile string
	SourceUrl string
	done <-chan error
}

func NewVideoImpl() *VideoImpl {
	pool := NewTaskPool(1024, runtime.NumCPU())
	return &VideoImpl{
		sync.RWMutex{},
		pool,
		1024,
		make(map[int]VideoSplitTask),
	}
}

func (v *VideoImpl) VideoWsInterface(ctx *gin.Context) {
	u := websocket.Upgrader{}
	conn, err := u.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		Logger.ErrorFromErr(err)
		return
	}
	defer conn.Close()
	tmp,ok := ctx.Get("UserInfo")
	if !ok {
		ctx.JSON(http.StatusOK,*ErrServer)
		return
	}
	userInfo := tmp.(UserInfo)
	for {
		msgTyp, buf, err := conn.ReadMessage()
		if err != nil {
			Logger.ErrorFromErr(err)
			return
		}
		if msgTyp == websocket.PingMessage {
			err := conn.WriteMessage(websocket.PongMessage, []byte("pong"))
			if err != nil {
				Logger.ErrorFromErr(err)
				return
			}
		}
		var req RequestFrame
		err = json.Unmarshal(buf, &req)
		if err != nil {
			Logger.ErrorFromErr(err)
			return
		}
		if req.Type == "check_status" {
			v.mu.Lock()
			task,ok := v.taskCollection[req.TaskId]
			if !ok {
				v.mu.Unlock()
				bytes,_ := json.Marshal(ErrVideoTask)
				err := conn.WriteMessage(websocket.TextMessage, bytes)
				if err != nil {
					Logger.ErrorFromErr(err)
				}
				continue
			}
			select {
			case err := <-task.done:
				delete(v.taskCollection,req.TaskId)
				v.mu.Unlock()
				if err != nil {
					bytes,_ := json.Marshal(ErrVideoTaskHandle)
					err := conn.WriteMessage(websocket.TextMessage, bytes)
					if err != nil {
						Logger.ErrorFromErr(err)
					}
					continue
				}
				bytes, _ := json.Marshal(*Ok)
				err = conn.WriteMessage(websocket.TextMessage, bytes)
				if err != nil {
					Logger.ErrorFromErr(err)
				}
			default:
				v.mu.Unlock()
				bytes, _ := json.Marshal(ErrVideoTaskNoOk)
				err := conn.WriteMessage(websocket.TextMessage, bytes)
				if err != nil {
					Logger.ErrorFromErr(err)
				}
			}
			continue
		}
		newUUID,err := uuid.NewUUID()
		if err != nil {
			ctx.JSON(http.StatusOK,*ErrServer)
			return
		}
		done := make(chan error)
		task := VideoSplitTask{
			SplitVideoFile: newUUID.String() + "." + req.VideoOutType,
			SourceUrl:      req.VideoUrl,
			done:           done,
		}
		v.mu.Lock()
		id := v.taskIdStart
		v.taskIdStart++
		v.taskCollection[id] = task
		v.mu.Unlock()
		ss,err := FFmpegTimeParse(req.VideoStart)
		if err != nil {
			bytes,_ := json.Marshal(ErrJsonParam)
			err := conn.WriteMessage(websocket.TextMessage, bytes)
			if err != nil {
				Logger.ErrorFromErr(err)
			}
			continue
		}
		end,err := FFmpegTimeParse(req.VideoEnd)
		if err != nil {
			bytes,_ := json.Marshal(ErrJsonParam)
			err := conn.WriteMessage(websocket.TextMessage, bytes)
			if err != nil {
				Logger.ErrorFromErr(err)
			}
			continue
		}
		et := time.Time{}.Add(end.Sub(ss))
		v.pool.Push(func() {
			err := GetVideoFromUrl(req.VideoUrl, task.SplitVideoFile)
			if err != nil {
				done <- err
				return
			}
			callEnd := FFmpegTimeFormat(et)
			err = CallFfmpeg(nil,"/tmp/" + task.SplitVideoFile,
				VIDEO_SOURCE_PATH + "/" + task.SplitVideoFile,
				req.VideoStart,callEnd)
			if err != nil {
				done <- err
				return
			}
			var record model.UserRecord
			record.Uid = userInfo.Uid
			record.VideoId = task.SplitVideoFile
			done <- record.Create()
		})
		var rep ResponseFrame
		rep.TaskId = id
		rep.StateComment = "start"
		rep.SplitTime = end.Sub(ss)
		rep.Type = "return"
		rep.SplitVideoId = task.SplitVideoFile
		bytes,err := json.Marshal(&rep)
		if err != nil {
			Logger.ErrorFromErr(err)
			return
		}
		err = conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			return
		}
	}
}

func (v *VideoImpl) VideoGetStatus(ctx *gin.Context) {
	v.mu.RLock()
	videoId := ctx.Param("videoId")
	id, err := strconv.Atoi(videoId)
	if err != nil {
		ctx.JSON(http.StatusOK,ErrJsonParam)
		return
	}
	task,ok := v.taskCollection[id]
	if !ok {
		v.mu.Unlock()
		ctx.JSON(http.StatusOK,*ErrVideoTask)
		return
	}
	select {
	case err := <-task.done:
		delete(v.taskCollection,id)
		v.mu.Unlock()
		if err != nil {
			ctx.JSON(http.StatusOK,ErrVideoTaskHandle)
			return
		}
		ctx.JSON(http.StatusOK,Ok)
	default:
		v.mu.Unlock()
		ctx.JSON(http.StatusOK,ErrVideoTaskNoOk)
	}
}


type UserInfo struct {
	Uid      int64
	UserName string
}

type UserModifyInfo struct {
	UserName    string
	OldPassword string
	NewPassword string
}

func UserPostStatement(ctx *gin.Context) {
	state := ctx.Param("state")
	switch state {
	case "sign":
		user := &model.User{}
		err := ctx.BindJSON(&user)
		if err != nil {
			ctx.JSON(http.StatusOK, ErrJsonParam)
			return
		}
		err = user.CreateUser(user)
		if err != nil {
			uerr := *ErrServer
			uerr.Msg = err.Error()
			ctx.JSON(http.StatusOK, uerr)
			return
		}
		ctx.JSON(http.StatusOK, *Ok)
	case "login":
		user := &model.User{}
		err := ctx.BindJSON(&user)
		if err != nil {
			ctx.JSON(http.StatusOK, ErrJsonParam)
			return
		}
		user, err = user.VerifyPassword(user.UserName, user.UserPassword)
		if err != nil {
			ctx.JSON(http.StatusOK, ServerErr{
				Code: -1,
				Msg:  err.Error(),
			})
		}
		newUUID, err := uuid.NewUUID()
		if err != nil {
			ctx.JSON(http.StatusOK, ErrServer)
			return
		}
		tokenPool.AddToken(newUUID.String(), UserInfo{
			Uid:      user.Uid,
			UserName: user.UserName,
		}, 90*time.Second)
		ctx.JSON(http.StatusOK, struct {
			ServerErr
			Token string `json:"token"`
		}{
			*Ok,
			newUUID.String(),
		})
	}
}

func UserModify(ctx *gin.Context) {
	tmp, _ := ctx.Get("UserInfo")
	userInfo := tmp.(UserInfo)
	var userModify UserModifyInfo
	err := ctx.BindJSON(&userModify)
	if err != nil {
		Logger.ErrorFromErr(err)
		ctx.JSON(http.StatusOK, ErrJsonParam)
		return
	}
	var user model.User
	err = user.SelectUser1(userInfo.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, ErrServer)
		return
	}
	if user.UserName == userModify.UserName && user.UserPassword == userModify.OldPassword {
		user.UserPassword = userModify.NewPassword
		err := user.ModifyUser(&user)
		if err != nil {
			ctx.JSON(http.StatusOK, ErrServer)
			return
		}
	} else {
		ctx.JSON(http.StatusOK, ErrServer)
	}
}

func UserDelete(ctx *gin.Context) {
	tmp, _ := ctx.Get("UserInfo")
	userInfo := tmp.(UserInfo)
	var user model.User
	user.Uid = userInfo.Uid
	user.UserName = userInfo.UserName
	err := user.DeleteUser(&user)
	if err != nil {
		Logger.ErrorFromErr(err)
		ctx.JSON(http.StatusOK, ErrServer)
	}
}

func UserGet(ctx *gin.Context) {
	uid, err := strconv.ParseInt(ctx.Param("uid"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ErrJsonParam)
		return
	}
	user := model.User{}
	err = user.SelectUser1(uid)
	if err != nil {
		ctx.JSON(http.StatusOK, ErrServer)
		return
	}
	ctx.JSON(http.StatusOK, struct {
		ServerErr
		Data interface{} `json:"data"`
	}{
		*Ok,
		UserInfo{
			Uid:      user.Uid,
			UserName: user.UserName,
		},
	})
}
