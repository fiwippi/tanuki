package user

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func GetType(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("uid")
		u, err := s.Store.GetUser(uid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, gin.H{"type": u.Type})
	}
}
