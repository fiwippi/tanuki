package auth

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// LogoutReply defines the reply from /api/auth/logout
type LogoutReply struct {
	Success bool `json:"success"`
}

// GET /api/auth/logout
func Logout(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		s.Session.Delete(c)
		uid := c.GetString("uid")
		log.Debug().Str("uid", uid).Msg("user logged out")
		c.JSON(200, LogoutReply{Success: true})
	}
}
