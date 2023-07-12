package mangadex

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/server"
)

func getCover(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		endpoint := c.Param("endpoint")
		if endpoint == "" {
			c.AbortWithStatus(400)
			return
		}
		endpoint = strings.ReplaceAll(endpoint, "_", "/")

		resp, err := mangadex.GetCover(c, endpoint)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		defer resp.Body.Close()

		reader := resp.Body
		contentLength := resp.ContentLength
		contentType := resp.Header.Get("Content-Type")
		c.DataFromReader(200, contentLength, contentType, reader, nil)
	}
}
