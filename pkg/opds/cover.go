package opds

import (
	admin "github.com/fiwippi/tanuki/pkg/api/series"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /opds/v1.2/series/:sid/entries/:eid/cover?thumbnail={true,false}
func GetCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		admin.GetEntryCover(s)(c)
	}
}
