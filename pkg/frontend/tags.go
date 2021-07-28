package frontend

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /tags/:tag
func specificTag(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		tag := c.Param("tag")

		allTags := s.Store.GetTags()
		if !allTags.Has(tag) {
			s.Err404(c)
			return
		}
		c.HTML(200, "specific-tag.tmpl", nil)
	}
}

// GET /tags
func tags(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "tags.tmpl", nil)
	}
}
