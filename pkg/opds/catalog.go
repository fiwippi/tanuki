package opds

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/feed"
	"github.com/fiwippi/tanuki/pkg/server"
)

func GetCatalog(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctl, err := s.Store.GetCatalog()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		var modTime time.Time
		for _, series := range ctl {
			t := series.ModTime.Time()
			if t.After(modTime) {
				modTime = t
			}
		}

		f := feed.NewCatalogFeed(opdsRoot)
		f.SetAuthor(authorName, authorURI)
		f.SetUpdated(modTime)
		for _, series := range ctl {
			f.AddSeries(series.SID, series.Title(), series.ModTime.Time())
		}

		c.XML(200, f)
	}
}
