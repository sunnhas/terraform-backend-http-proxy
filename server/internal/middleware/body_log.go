package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"terraform-backend-http-proxy/server/internal/ginutils"
)

func BodyLog(c *gin.Context) {
	rawJson, _ := ginutils.GetBody(c)
	log.Printf("Body: %s\n", string(rawJson))
}
