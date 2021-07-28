package frontend

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

func Err404(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(404, "404.tmpl", nil)
	}
}
