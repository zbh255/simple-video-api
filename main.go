package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zbh255/video-api/model"
)

func main() {
	eng := gin.Default()
	model.InitOrm("./testdata/test.db")
	InitRouter(eng)
	err := eng.Run(":1234")
	Logger.Debug(fmt.Sprintf("gin server close : %s",err.Error()))
}

