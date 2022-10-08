package server

import (
	"github.com/gin-gonic/gin"
	"terraform-backend-http-proxy/server/internal/handler"
	"terraform-backend-http-proxy/server/internal/middleware"
)

func Run() {
	r := gin.Default()

	r.Use(middleware.BodyLog)
	r.Use(middleware.ParseRequestData)

	r.GET("/", handler.GetState)
	r.POST("/", handler.UpdateState)
	r.DELETE("/", handler.DeleteState)
	r.Handle("LOCK", "/", handler.LockState)
	r.Handle("UNLOCK", "/", handler.UnlockState)

	err := r.Run("localhost:6061")
	if err != nil {
		panic(err)
	}
}
