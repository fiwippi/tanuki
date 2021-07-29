package frontend

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /
func home(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "home.tmpl", c)
	}
}
