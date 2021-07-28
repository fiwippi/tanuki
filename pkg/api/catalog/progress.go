package catalog

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ProgressReply for /api/catalog/progress
type ProgressReply struct {
	Success  bool                             `json:"success"`
	Progress map[string]*users.SeriesProgress `json:"progress"`
}

// GET /api/catalog/progress
func GetProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("uid")

		user, err := s.Store.GetUser(uid)
		if err != nil {
			log.Debug().Err(err).Str("uid", uid).Msg("could not get user")
			c.AbortWithStatusJSON(500, ProgressReply{Success: false})
			return
		}

		// Return the progress
		c.JSON(200, ProgressReply{Success: true, Progress: user.Progress.Data})
	}
}
