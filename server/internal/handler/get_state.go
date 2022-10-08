package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"terraform-backend-http-proxy/backend"
	"terraform-backend-http-proxy/server/internal/ginutils"
	"terraform-backend-http-proxy/server/internal/middleware"
)

func GetState(c *gin.Context) {
	requestData := middleware.ReadRequestData(c)

	state, err := backend.GetState(requestData)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Status(http.StatusNoContent)
			return
		}

		ginutils.ServerError(c, err)
		return
	}

	c.Data(http.StatusOK, "application/json", state)
}
