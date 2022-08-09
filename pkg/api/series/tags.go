package series

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/manga"
	"github.com/fiwippi/tanuki/pkg/server"
)

func PatchSeriesTags(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			Tags *manga.Tags `json:"tags"`
		}{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithError(400, err)
			return
		}

		sid := c.Param("sid")
		if err := s.Store.SetSeriesTags(sid, data.Tags); err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Status(200)
	}
}
