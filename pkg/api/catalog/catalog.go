package catalog

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/gin-gonic/gin"
)

// CatalogReply for /api/series
type CatalogReply struct {
	Success bool        `json:"success"`
	List    api.Catalog `json:"list"`
}

// GET /api/series
func GetCatalog(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		list := s.Store.GetCatalog()
		c.JSON(200, CatalogReply{Success: true, List: list})
	}
}
