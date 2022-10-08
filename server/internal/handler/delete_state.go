package handler

import (
	"github.com/gin-gonic/gin"
	"terraform-backend-http-proxy/backend"
)

func DeleteState(c *gin.Context) {
	backend.DeleteState()
}
