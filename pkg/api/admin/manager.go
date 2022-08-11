package admin

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func checkSubscription(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		err := s.Manager.CheckSubscriptions()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, gin.H{"time_taken": time.Now().Sub(now).String()})
	}
}
