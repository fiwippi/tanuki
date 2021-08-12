package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func Err404(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(404, "404.tmpl", nil)
	}
}
