package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/pkg/server"
)

func login(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the request
		data := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatus(400)
			return
		}

		// Validate login details
		valid := s.Store.ValidateLogin(data.Username, data.Password)
		if !valid {
			c.AbortWithStatus(401)
			return
		}
		log.Debug().Str("username", data.Username).Msg("validated user")

		// If valid then give user token they can identify themselves with
		uid := hash.SHA1(data.Username)
		s.Session.Store(uid, c)
		c.Status(200)
	}
}
