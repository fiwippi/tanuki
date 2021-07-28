package admin

import (
	"fmt"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"time"
)

// LibraryScanReply for /api/admin/library
type LibraryScanReply struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// LibraryMissingEntriesReply for /api/admin/library/missing-items
type LibraryMissingEntriesReply struct {
	Success bool             `json:"success"`
	Entries api.MissingItems `json:"entries"`
}

// GET /api/admin/library/scan
func ScanLibrary(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		err := s.ScanLibrary()
		if err != nil {
			log.Error().Err(err).Msg("failed to scan library")
			c.AbortWithStatusJSON(500, LibraryScanReply{Success: false})
			return
		} else {
			timeTaken := time.Now().Sub(now)
			c.JSON(200, LibraryScanReply{Success: true, Message: fmt.Sprintf("The time taken was %s", timeTaken)})
		}
	}
}

// GET /api/admin/library/generate-thumbnails
func GenerateThumbnails(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		err := s.Store.GenerateThumbnails(true)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate thumbnails")
			c.AbortWithStatusJSON(500, LibraryScanReply{Success: false})
		} else {
			timeTaken := time.Now().Sub(now)
			c.JSON(200, LibraryScanReply{Success: true, Message: fmt.Sprintf("The time taken was %s", timeTaken)})
		}
	}
}

// GET /api/admin/library/missing-items
func GetMissingItems(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, LibraryMissingEntriesReply{Success: true, Entries: s.Store.GetMissingItems()})
	}
}

// DELETE /api/admin/library/missing-items
func DeleteMissingItems(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := s.Store.DeleteMissingItems()
		if err != nil {
			log.Debug().Err(err).Msg("failed to delete missing items")
			c.AbortWithStatusJSON(500, LibraryMissingEntriesReply{Success: false})
			return
		}

		c.JSON(200, LibraryMissingEntriesReply{Success: true})
	}
}
