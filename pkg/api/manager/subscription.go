package manager

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func deleteSubscription(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := s.Store.DeleteSubscription(c.Param("sid"))
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Status(200)
	}
}

func getAllSubscriptions(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sub, err := s.Store.GetAllSubscriptions()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, sub)
	}
}
