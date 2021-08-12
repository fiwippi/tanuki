package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/server"
)

// GET /opds/v1.2/series/:sid
func GetEntries(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")

		data, err := s.Store.GetSeries(id)
		if err != nil {
			c.AbortWithStatus(404)
			return
		}
		series := feed.NewSeries(data.Hash, data.Title)
		series.SetAuthor(s.Author)
		updated, err := s.Store.GetSeriesModTime(id)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		series.SetUpdated(updated)

		list, err := s.Store.GetEntries(id)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		for _, e := range list {
			entryTime, err := s.Store.GetEntryModTime(id, e.Hash)
			if err != nil {
				c.AbortWithStatus(500)
				return
			}
			cover, err := s.Store.GetEntryCover(id, e.Hash)
			if err != nil {
				c.AbortWithStatus(500)
				return
			}
			a, err := s.Store.GetEntryArchive(id, e.Hash)
			if err != nil {
				c.AbortWithStatus(500)
				return
			}
			p, err := s.Store.GetEntryPage(id, e.Hash, 1)
			if err != nil {
				c.AbortWithStatus(500)
				return
			}

			series.AddEntry(&feed.ArchiveEntry{
				Title:     e.Title,
				Updated:   feed.Time{Time: entryTime},
				ID:        e.Hash,
				CoverType: cover.ImageType,
				ThumbType: image.JPEG,
				PageType:  p.ImageType,
				Archive:   a,
				Pages:     e.Pages,
			})
		}

		c.XML(200, series)
	}
}
