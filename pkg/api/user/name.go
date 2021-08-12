package user

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

// NameReply defines the reply from /api/user/name
type NameReply struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
}

// GET /api/user/name
func GetName(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("uid")
		u, err := s.Store.GetUser(uid)
		if err != nil {
			c.AbortWithStatusJSON(500, NameReply{Success: false})
			return
		}

		c.JSON(200, NameReply{Success: true, Name: u.Name})
	}
}
