package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func login(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "login.tmpl", nil)
	}
}
