package admin

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func ScanLibrary(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		err := s.Store.PopulateCatalog()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, gin.H{"time_taken": time.Since(now).String()})
	}
}

func GenerateThumbnails(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		err := s.Store.GenerateThumbnails(true)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, gin.H{"time_taken": time.Since(now).String()})
	}
}

func GetMissingItems(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := s.Store.GetMissingItems()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, gin.H{"items": items})
	}
}

func DeleteMissingItems(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := s.Store.DeleteMissingItems()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Status(200)
	}
}
