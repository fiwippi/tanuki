package download

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/server"
)

// ChaptersRequest defines the request to /api/download/chapters
type ChaptersRequest struct {
	Title    string            `json:"title"`
	Chapters mangadex.Chapters `json:"chapters"`
}

// ChaptersReply defines the replies from /api/download/chapters
type ChaptersReply struct {
	Success bool `json:"success"`
}

func GetChapters(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the request
		var data ChaptersRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatusJSON(400, ChaptersReply{Success: false})
			return
		}
		c.JSON(200, ChaptersReply{Success: true})

		// Start downloading the chapters
		for _, ch := range data.Chapters {
			go manager.StartDownload(data.Title, ch)
		}
	}
}
