package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func reader(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		_, err := s.Store.GetEntry(sid, eid)
		if err != nil {
			s.Err404(c)
			return
		}

		c.HTML(200, "reader.tmpl", c)
	}
}
