package series

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

// TagsRequest for the route /api/tag/:id/series
type TagsRequest struct {
	Tags []string `json:"tags"`
}

// TagsReply for the route /api/tag/:id/series
type TagsReply struct {
	Success bool `json:"success"`
}

func PatchSeriesTags(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data TagsRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatusJSON(400, TagsReply{Success: false})
			return
		}

		id := c.Param("sid")
		if err := s.Store.SetSeriesTags(id, data.Tags); err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, TagsReply{Success: true})
	}
}
