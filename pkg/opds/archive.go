package opds

import (
	"github.com/fiwippi/tanuki/pkg/api/series"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /opds/v1.2/series/:sid/entries/:eid/archive
func GetArchive(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		series.GetEntryArchive(s)(c)
	}
}