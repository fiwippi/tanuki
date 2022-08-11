package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/feed"
	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/pkg/server"
)

func GetEntries(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")

		series, err := s.Store.GetSeries(sid)
		if err != nil {
			c.AbortWithStatus(404)
			return
		}

		f := feed.NewSeriesFeed(opdsRoot, series.SID, series.Title())
		f.SetAuthor(authorName, authorURI)
		f.SetUpdated(series.ModTime.Time())

		entries, err := s.Store.GetEntries(sid)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		for _, e := range entries {
			tt := image.JPEG
			pt := image.Invalid
			if len(e.Pages) > 0 {
				pt = e.Pages[0].Type
			}
			ct, err := s.Store.GetEntryCoverType(sid, e.EID)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			if ct == image.Invalid {
				ct = pt
			}

			f.AddEntry(e.EID, e.Title(), tt.MimeType(), ct.MimeType(),
				pt.MimeType(), len(e.Pages), e.ModTime.Time(), &e.Archive)
		}

		c.XML(200, f)
	}
}
