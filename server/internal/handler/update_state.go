package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"terraform-backend-http-proxy/backend"
	ginutils2 "terraform-backend-http-proxy/server/internal/ginutils"
	"terraform-backend-http-proxy/server/internal/middleware"
)

func UpdateState(c *gin.Context) {
	requestData := middleware.ReadRequestData(c)

	body, err := ginutils2.GetBody(c)
	if err != nil {
		ginutils2.ServerError(c, err)
		return
	}

	if err := backend.UpdateState(requestData, body); err != nil {
		ginutils2.ServerError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
