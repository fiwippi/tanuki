package admin

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

// LibraryScanReply for /api/admin/library
type LibraryScanReply struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// LibraryMissingEntriesReply for /api/admin/library/missing-items
type LibraryMissingEntriesReply struct {
	Success bool             `json:"success"`
	Items   api.MissingItems `json:"items"`
}

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

func GetMissingItems(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, LibraryMissingEntriesReply{Success: true, Items: s.Store.GetMissingItems()})
	}
}

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
