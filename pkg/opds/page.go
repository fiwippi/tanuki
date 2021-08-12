package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/api/series"
	"github.com/fiwippi/tanuki/pkg/server"
)

// GET /opds/v1.2/series/:sid/entries/:eid/page/:num?zero_based={true|false}
func GetPage(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		series.GetEntryPage(s)(c)
	}
}
