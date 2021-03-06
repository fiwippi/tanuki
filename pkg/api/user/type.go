package user

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
)

// TypeReply defines the reply from /api/user/type
type TypeReply struct {
	Success bool       `json:"success"`
	Type    users.Type `json:"type"`
}

func GetType(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("uid")
		u, err := s.Store.GetUser(uid)
		if err != nil {
			c.AbortWithStatusJSON(500, TypeReply{Success: false})
			return
		}

		c.JSON(200, TypeReply{Success: true, Type: u.Type})
	}
}
