package tanuki

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/api"
)

// GET /api/tags
func apiGetAllTags(c *gin.Context) {
	t := db.GetTags()
	c.JSON(200, api.AllTagsReply{Success: true, Tags: t.List()})
}

// GET /api/tag/:tag
func apiGetSeriesWithTag(c *gin.Context) {
	tag := c.Param("tag")
	t := db.GetSeriesWithTag(tag)
	c.JSON(200, api.SeriesListReply{Success: true, List: t})
}