package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func UserAuthHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenStr := context.GetHeader("Authorization")
		authStrs := strings.SplitN(tokenStr," ",2)
		if len(authStrs) != 2 {
			context.Abort()
			context.JSON(http.StatusOK,ErrJsonParam)
			return
		}
		tokenStr = authStrs[1]
		token,ok := tokenPool.GetToken(tokenStr)
		if !ok {
			context.Abort()
			context.JSON(http.StatusOK,ErrUserSign)
			return
		}
		userInfo := token.Data.(UserInfo)
		context.Set("UserInfo", userInfo)
	}
}
