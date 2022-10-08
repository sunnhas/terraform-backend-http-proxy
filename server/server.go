package server

import (
	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()

	err := r.Run("localhost:6061")
	if err != nil {
		panic(err)
	}
}
