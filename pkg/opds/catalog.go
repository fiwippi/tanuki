package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/fuzzy"
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/server"
)

func getCatalog(s *server.Server, filter string) (*feed.Catalog, error) {
	catalog := feed.NewCatalog()
	catalog.SetAuthor(s.Author)

	updated, err := s.Store.GetCatalogModTime()
	if err != nil {
		return nil, err
	}
	catalog.SetUpdated(updated)

	list := s.Store.GetCatalog()
	for _, series := range list {
		if len(filter) > 0 && !fuzzy.Search(series.Title, filter) {
			continue
		}

		seriesTime, err := s.Store.GetSeriesModTime(series.Hash)
		if err != nil {
			return nil, err
		}
		catalog.AddEntry(&feed.SeriesEntry{
			Title:   series.Title,
			Updated: feed.Time{Time: seriesTime},
			ID:      series.Hash,
		})
	}

	return catalog, nil
}

// GET /opds/v1.2/catalog
func GetCatalog(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		catalog, err := getCatalog(s, c.Query("search"))
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		c.XML(200, catalog)
	}
}
