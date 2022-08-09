package series

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/pkg/server"
)

func GetSeriesCover(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")
		thumbnail := c.DefaultQuery("thumbnail", "false")

		var img []byte
		var err error
		var imType image.Type
		if thumbnail == "true" {
			img, imType, err = s.Store.GetSeriesThumbnail(id)
		} else {
			img, imType, err = s.Store.GetSeriesCover(id)
		}

		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		if len(img) == 0 {
			c.AbortWithError(500, fmt.Errorf("img data is empty, thumbnail: %s", thumbnail))
			return
		}

		c.Data(200, imType.MimeType(), img)
	}
}

func GetEntryCover(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		thumbnail := c.DefaultQuery("thumbnail", "false")

		var img []byte
		var err error
		var imType image.Type
		if thumbnail == "true" {
			img, imType, err = s.Store.GetEntryThumbnail(sid, eid)
		} else {
			img, imType, err = s.Store.GetEntryCover(sid, eid)
		}

		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Data(200, imType.MimeType(), img)
	}
}
