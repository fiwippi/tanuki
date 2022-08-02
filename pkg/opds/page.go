package opds

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func GetPage(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		numStr := c.Param("num")

		num, err := strconv.Atoi(numStr)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		var zb bool
		zbQuery := c.Query("zero_based")
		if len(zbQuery) > 0 {
			zb, err = strconv.ParseBool(zbQuery)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}
		}

		r, size, im, err := s.Store.GetPage(sid, eid, num, zb)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.DataFromReader(http.StatusOK, size, im.MimeType(), r, nil)
	}
}
