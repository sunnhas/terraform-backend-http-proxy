package ginutils

import "github.com/gin-gonic/gin"

const bodyReaderKey = "body-reader-key"

func GetBody(c *gin.Context) ([]byte, error) {
	var body []byte

	if cb, ok := c.Get(bodyReaderKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}

	if body == nil {
		b, err := c.GetRawData()
		if err != nil {
			return nil, err
		}
		body = b
		c.Set(bodyReaderKey, body)
	}

	return body, nil
}
