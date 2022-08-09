package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func entries(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")
		_, err := s.Store.GetSeries(id)
		if err != nil {
			c.AbortWithError(404, err)
			return
		}
		c.HTML(200, "entries.tmpl", c)
	}
}
