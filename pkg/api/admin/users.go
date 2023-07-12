package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/user"
)

func GetUsers(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		us, err := s.Store.GetUsers()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		for i := range us {
			us[i].Pass = ""
		}

		c.JSON(200, gin.H{"users": us})
	}
}

func PutUsers(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			Username string    `json:"username"`
			Password string    `json:"password"`
			Type     user.Type `json:"type"`
		}{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithStatus(400)
			return
		}

		if data.Username == "" {
			c.AbortWithStatusJSON(400, gin.H{"message": "username cannot be empty"})
			return
		}
		if len(data.Password) < 8 {
			c.AbortWithStatusJSON(400, gin.H{"message": "password should be minimum of 8 characters"})
			return
		}

		err := s.Store.AddUser(user.NewAccount(data.Username, data.Password, data.Type), false)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Status(200)
	}
}
