package download

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/downloading"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

// How often to update the user with the manager's status
var updateInterval = 1 * time.Second

// The global download manager
var manager *downloading.Manager

// ManagerStatusReply represents the current manager state
type ManagerStatusReply struct {
	Downloads []*api.Download `json:"downloads"`
	Paused    bool            `json:"paused"`
}

// ManagerChangeReply for when the user tries to change
// the manager state
type ManagerChangeReply struct {
	Success bool `json:"success"`
}

func ViewManager(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		c.Stream(func(w io.Writer) bool {
			<-ticker.C

			// The downloads slice comes from a sync.Pool, doneFunc() frees this memory
			dls, doneFunc := manager.Downloads()
			defer doneFunc()

			c.SSEvent("message", ManagerStatusReply{Downloads: dls, Paused: manager.Paused()})
			return true
		})
	}
}

func DeleteFinishedTasks(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := manager.DeleteFinishedTasks()
		if err != nil {
			c.AbortWithStatusJSON(500, ManagerChangeReply{Success: false})
			return
		}
		c.JSON(200, ManagerChangeReply{Success: true})
	}
}

func RetryFailedTasks(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		manager.RetryFailedTasks()
		c.JSON(200, ManagerChangeReply{Success: true})
	}
}

func PauseDownloads(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		manager.Pause()
		c.JSON(200, ManagerChangeReply{Success: true})
	}
}

func ResumeDownloads(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		manager.Resume()
		c.JSON(200, ManagerChangeReply{Success: true})
	}
}

func CancelDownloads(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		manager.Cancel()
		c.JSON(200, ManagerChangeReply{Success: true})
	}
}
