package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"terraform-backend-http-proxy/backend"
	ginutils2 "terraform-backend-http-proxy/server/internal/ginutils"
	"terraform-backend-http-proxy/server/internal/middleware"
)

func LockState(c *gin.Context) {
	requestData := middleware.ReadRequestData(c)

	body, err := ginutils2.GetBody(c)
	if err != nil {
		ginutils2.ServerError(c, err)
		return
	}

	if lockInfo, err := backend.LockState(requestData, body); err != nil {
		if errors.Is(err, backend.StateIsLocked) {
			c.JSON(http.StatusLocked, lockInfo)
			return
		}

		ginutils2.ServerError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
