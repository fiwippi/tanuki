package mangadex

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/server"
)

func searchMangadex(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		title := c.Query("title")
		if title == "" {
			c.AbortWithStatus(400)
			return
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "15"))
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ls, err := mangadex.SearchManga(ctx, title, limit)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, ls)
	}
}
