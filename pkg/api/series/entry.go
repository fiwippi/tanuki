package series

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func GetEntryArchive(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		e, err := s.Store.GetEntry(sid, eid)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		fp := fmt.Sprintf("%s.%s", e.Archive.Title, e.Archive.Type.String())
		c.FileAttachment(e.Archive.Path, fp)
	}
}

func GetEntryPage(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		numStr := c.Param("num")

		num, err := strconv.Atoi(numStr)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		var zb bool
		zbQuery := c.Query("zero_based")
		if len(zbQuery) > 0 {
			zb, err = strconv.ParseBool(zbQuery)
			if err != nil {
				c.AbortWithError(400, err)
				return
			}
		}

		r, size, im, err := s.Store.GetPage(sid, eid, num, zb)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.DataFromReader(200, size, im.MimeType(), r, nil)
	}
}
