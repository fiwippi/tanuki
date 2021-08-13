package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
)

// UsersPutRequest for /api/admin/users
type UsersPutRequest struct {
	Username string     `json:"username"`
	Password string     `json:"password"`
	Type     users.Type `json:"type"`
}

// UsersReply for /api/admin/users
type UsersReply struct {
	Success bool         `json:"success"`
	Users   []users.User `json:"users,omitempty"`
	Message string       `json:"message,omitempty"`
}

func GetUsers(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, UsersReply{Success: true, Users: s.Store.GetUsers(true)})
	}
}

func PutUsers(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data UsersPutRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
			c.AbortWithStatusJSON(400, UsersReply{Success: false})
			return
		}

		if data.Username == "" {
			c.AbortWithStatusJSON(400, UsersReply{Success: false, Message: "username cannot be empty"})
			return
		}

		if len(data.Password) < 8 {
			c.AbortWithStatusJSON(400, UsersReply{Success: false, Message: "password should be minimum of 8 characters"})
			return
		}

		err := s.Store.CreateUser(users.NewUser(data.Username, data.Password, data.Type))
		if err != nil {
			c.AbortWithStatusJSON(500, UsersReply{Success: false})
			return
		}

		c.JSON(200, UsersReply{Success: true})
	}
}
