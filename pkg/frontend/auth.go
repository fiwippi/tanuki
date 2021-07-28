package frontend

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /login
func login(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "login.tmpl", nil)
	}
}
