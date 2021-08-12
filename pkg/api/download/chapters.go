package download

import (
	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

type ChaptersReply struct {
	Success bool `json:"success"`
}

// GET /api/download/chapters
func GetChapters(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the request
		var data struct {
			Title    string            `json:"title"`
			Chapters mangadex.Chapters `json:"chapters"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatusJSON(400, ChaptersReply{Success: false})
			return
		}
		c.JSON(200, ChaptersReply{Success: true})

		for _, ch := range data.Chapters {
			go manager.StartDownload(data.Title, ch)
		}
	}
}
