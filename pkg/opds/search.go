package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/server"
)

func GetSearch(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := feed.NewDefaultSearch()
		s.URL.Template = "/opds/v1.2/catalog?search={searchTerms}"
		c.XML(200, s)
	}
}
