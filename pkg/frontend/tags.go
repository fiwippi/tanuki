package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func specificTag(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		tag := c.Param("tag")

		allTags, err := s.Store.GetAllTags()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		if !allTags.Has(tag) {
			c.AbortWithStatus(404)
			return
		}
		c.HTML(200, "specific-tag.tmpl", c)
	}
}

func tags(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "tags.tmpl", c)
	}
}
