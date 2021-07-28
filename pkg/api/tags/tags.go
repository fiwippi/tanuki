package tags

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/gin-gonic/gin"
)

// TagsReply for the route /api/tags
type TagsReply struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
}

// SeriesTagsRequest for the route /api/tag/:id/series
type SeriesTagsRequest struct {
	Tags []string `json:"tags"`
}

// SeriesTagsReply for the route /api/tag/:id/series
type SeriesTagsReply struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
}

//
type SeriesWithTagReply struct {
	Success bool        `json:"success"`
	List    api.Catalog `json:"list"`
}

// GET /api/tags
func GetTags(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := s.Store.GetTags()
		c.JSON(200, TagsReply{Success: true, Tags: t.List()})
	}
}

// GET /api/tag/:tag
func GetSeriesWithTag(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		tag := c.Param("tag")
		t := s.Store.GetSeriesWithTag(tag)
		c.JSON(200, SeriesWithTagReply{Success: true, List: t})
	}
}
