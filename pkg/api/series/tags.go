package series

import (
	"github.com/fiwippi/tanuki/pkg/api/tags"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GET /api/series/:sid/tags
func GetSeriesTags(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")
		t, err := s.Store.GetSeriesTags(id)
		if err != nil {
			c.AbortWithStatusJSON(500, tags.SeriesTagsReply{Success: false})
			return
		}
		c.JSON(200, tags.SeriesTagsReply{Success: true, Tags: t.List()})
	}
}

// PATCH /api/series/:sid/tags
func PatchSeriesTags(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data tags.SeriesTagsRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
			c.AbortWithStatusJSON(400, tags.SeriesTagsReply{Success: false})
			return
		}

		id := c.Param("sid")
		if err := s.Store.SetSeriesTags(id, data.Tags); err != nil {
			log.Debug().Err(err).Str("series", id).Msg("failed to set tags")
			c.AbortWithStatusJSON(500, tags.SeriesTagsReply{Success: false})
			return
		}

		c.JSON(200, tags.SeriesTagsReply{Success: true})
	}
}
