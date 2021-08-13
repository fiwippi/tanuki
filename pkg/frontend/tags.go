package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func specificTag(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		tag := c.Param("tag")

		allTags := s.Store.GetTags()
		if !allTags.Has(tag) {
			s.Err404(c)
			return
		}
		c.HTML(200, "specific-tag.tmpl", c)
	}
}

func tags(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "tags.tmpl", c)
	}
}
