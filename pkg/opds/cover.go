package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/api/series"
	"github.com/fiwippi/tanuki/pkg/server"
)

// GET /opds/v1.2/series/:sid/entries/:eid/cover?thumbnail={true,false}
func GetCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		series.GetEntryCover(s)(c)
	}
}
