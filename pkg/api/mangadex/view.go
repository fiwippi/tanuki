package mangadex

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/server"
)

func viewManga(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		if uuid == "" {
			c.AbortWithStatus(400)
			return
		}

		l, err := mangadex.ViewManga(context.Background(), uuid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		chs, err := l.ListChapters(context.Background())
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, gin.H{"listing": l, "chapters": chs})
	}
}
