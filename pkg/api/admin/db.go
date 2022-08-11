package admin

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func ViewStore(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Disposition", "attachment; filename=\"store.txt\"")
		d, err := s.Store.Dump()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Data(200, "text/plain", []byte(d))
	}
}

func VacuumStore(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		err := s.Store.Vacuum()
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, gin.H{"time_taken": time.Now().Sub(now).String()})
	}
}
