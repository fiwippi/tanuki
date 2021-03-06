package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func entries(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")
		_, err := s.Store.GetSeries(id)
		if err != nil {
			s.Err404(c)
			return
		}
		c.HTML(200, "entries.tmpl", c)
	}
}
