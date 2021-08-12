package download

import (
	"github.com/fiwippi/tanuki/pkg/downloading"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

var manager *downloading.Manager

type ManagerStatusReply struct {
	Downloads []*api.Download `json:"downloads"`
	Paused    bool            `json:"paused"`
}

type ManagerChangeReply struct {
	Success bool `json:"success"`
}

func ViewManager(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		c.Stream(func(w io.Writer) bool {
			<-ticker.C
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
