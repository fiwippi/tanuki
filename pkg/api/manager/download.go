package manager

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/server"
)

// How often to update the user with the manager's status
var updateInterval = 1 * time.Second

func viewDownloads(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		c.Stream(func(w io.Writer) bool {
			<-ticker.C

			// The downloads slice comes from a sync.Pool, doneFunc() frees this memory
			dls, doneFunc, err := s.Manager.GetAllDownloads()
			if err != nil {
				c.AbortWithError(500, err)
				return false
			}
			defer doneFunc()

			c.SSEvent("message", gin.H{
				"downloads": dls,
				"paused":    s.Manager.Paused(),
				"waiting":   s.Manager.Waiting(),
			})
			return true
		})
	}
}

func downloadChapters(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			Title              string             `json:"title"`
			Chapters           []mangadex.Chapter `json:"chapters"`
			CreateSubscription bool               `json:"create_subscription"`
		}{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatus(400)
			return
		}

		for _, ch := range data.Chapters {
			go s.Manager.Queue(data.Title, ch, data.CreateSubscription)
		}

		c.Status(200)
	}
}
