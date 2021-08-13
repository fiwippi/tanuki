package proxy

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/server"
)

func mangadex(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data ProxyRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatus(400)
			return
		}

		if len(data.Endpoint) == 0 {
			log.Debug().Interface("request", data).Msg("invalid proxy request")
			return
		}

		v, err := url.ParseQuery(data.Query)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		resp, err := s.Mangadex.Request("GET", fmt.Sprintf("%s?%s", data.Endpoint, v.Encode()), nil)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		if resp.StatusCode != 200 {
			c.Status(resp.StatusCode)
		} else {
			c.DataFromReader(200, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
		}
	}
}
