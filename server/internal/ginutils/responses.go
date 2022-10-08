package ginutils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func ServerError(c *gin.Context, err error) {
	//log.Println(err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "InternalServerError",
		"message": err,
	})
	panic(err)
}
