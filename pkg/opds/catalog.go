package opds

import (
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /opds/v1.2/catalog
func GetCatalog(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		catalog := feed.NewCatalog()
		catalog.SetAuthor(s.Author)

		updated, err := s.Store.GetCatalogModTime()
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		catalog.SetUpdated(updated)

		list := s.Store.GetCatalog()
		for _, series := range list {
			seriesTime, err := s.Store.GetSeriesModTime(series.Hash)
			if err != nil {
				c.AbortWithStatus(500)
				return
			}
			catalog.AddEntry(&feed.SeriesEntry{
				Title:   series.Title,
				Updated: feed.Time{Time: seriesTime},
				ID:      series.Hash,
			})
		}

		c.XML(200, catalog)
	}
}
