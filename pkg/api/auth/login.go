package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/pkg/server"
)

// We store the uid in the session, it becomes more
// efficient to access the database and improve security

// LoginRequest defines the request to /api/auth/login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginReply defines the reply from /api/auth/login
type LoginReply struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func Login(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the request
		var data LoginRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatusJSON(400, LoginReply{Success: false, Message: ""})
			return
		}

		// Validate login details
		valid := s.Store.ValidateLogin(data.Username, data.Password)
		if !valid {
			c.AbortWithStatusJSON(401, LoginReply{Success: false, Message: "invalid"})
			return
		}
		log.Debug().Str("username", data.Username).Msg("validated user")

		// If valid then give user token they can identify themselves with
		uid := hash.SHA1(data.Username)
		s.Session.Store(uid, c)
		c.JSON(200, LoginReply{Success: true})
	}
}
