package middleware

import (
	"github.com/gin-gonic/gin"
	"terraform-backend-http-proxy/backend"
	"terraform-backend-http-proxy/server/internal/ginutils"
	"terraform-backend-http-proxy/storage/storagetypes"
)

const requestDataKey = "request-data"

// ParseRequestData is a middleware parsing the request data in backend
// and storing it in the gin.Context of the request.
func ParseRequestData(c *gin.Context) {
	requestData, err := backend.ParseRequestData(c)
	if err != nil {
		ginutils.ServerError(c, err)
		return
	}
	c.Set(requestDataKey, requestData)
}

// ReadRequestData is a helper ginutils to read the parsed data
// from the gin.Context of the request within handlers.
func ReadRequestData(c *gin.Context) *storagetypes.ClientData {
	if requestData, ok := c.Get(requestDataKey); ok {
		return requestData.(*storagetypes.ClientData)
	}

	return nil
}
