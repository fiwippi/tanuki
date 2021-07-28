package admin

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /api/admin/db
func GetDB(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=\"db.txt\"")
		c.Data(200, "text/plain", []byte(s.Store.String()))
	}
}
