package proxy

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/pkg/server"
)

var ErrInvalidProxyRequest = errors.New("invalid proxy request")

func mangadex(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data ProxyRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatus(400)
			return
		}

		if len(data.Endpoint) == 0 {
			c.AbortWithError(400, ErrInvalidProxyRequest.Fmt(data))
			return
		}

		v, err := url.ParseQuery(data.Query)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		resp, err := s.Mangadex.Request("GET", fmt.Sprintf("%s?%s", data.Endpoint, v.Encode()), nil)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	}
}
