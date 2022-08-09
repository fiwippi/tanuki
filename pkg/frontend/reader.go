package frontend

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func reader(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		has, err := s.Store.HasEntry(sid, eid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		if !has {
			c.AbortWithError(404, fmt.Errorf("entry does not exist"))
			return
		}

		c.HTML(200, "reader.tmpl", c)
	}
}
