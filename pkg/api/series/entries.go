package series

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

// SeriesEntriesReply for /api/series/:id/entries
type SeriesEntriesReply struct {
	Success bool        `json:"success"`
	List    api.Entries `json:"list"`
}

func GetEntries(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		entries, err := s.Store.GetEntries(sid)
		if err != nil {
			c.AbortWithStatusJSON(500, SeriesEntriesReply{Success: false})
			return
		}
		c.JSON(200, SeriesEntriesReply{Success: true, List: entries})
	}
}
