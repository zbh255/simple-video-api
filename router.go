package main

import (
	"github.com/gin-gonic/gin"
)

const (
	VIDEO_SOURCE_PATH = "./testdata/video"
)

func InitRouter(eng *gin.Engine) {
	// WebSocket Interface
	v := NewVideoImpl()
	eng.Group("/").
		Use(UserAuthHandler()).
		GET("/video",v.VideoWsInterface).
		GET("/video/:videoId",v.VideoGetStatus)
	// video resource
	// Url Example: /video/source/userId/video-split-id
	eng.Static("/video/source",VIDEO_SOURCE_PATH)
	// state == login/sign
	eng.POST("/user/:state",UserPostStatement)
	// name == userName
	eng.Group("/").
		Use(UserAuthHandler()).
		PUT("/user",UserModify).
		DELETE("/user",UserDelete).
		GET("/user/:uid",UserGet)
}
