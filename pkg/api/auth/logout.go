package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/server"
)

func logout(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		s.Session.Delete(c)
		uid := c.GetString("uid")
		log.Debug().Str("uid", uid).Msg("user logged out")
		c.Status(200)
	}
}
