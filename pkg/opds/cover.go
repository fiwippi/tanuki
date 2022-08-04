package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/pkg/server"
)

func GetCover(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		thumbnail := c.DefaultQuery("thumbnail", "false")

		var data []byte
		var err error
		var img image.Type
		if thumbnail == "true" {
			data, img, err = s.Store.GetEntryThumbnail(sid, eid)
		} else {
			data, img, err = s.Store.GetEntryCover(sid, eid)
		}

		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Data(200, img.MimeType(), data)
	}
}
