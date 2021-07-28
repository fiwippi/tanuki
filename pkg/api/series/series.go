package series

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// SeriesReply for /api/series/:sid
type SeriesReply struct {
	Success bool       `json:"success"`
	Data    api.Series `json:"data"`
}

// GET /api/series/:sid
func GetSeries(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")
		s, err := s.Store.GetSeries(id)
		if err != nil {
			c.AbortWithStatusJSON(500, SeriesReply{Success: false})
			return
		}
		c.JSON(200, SeriesReply{Success: true, Data: *s})
	}
}

// PATCH /api/series/:sid
func PatchSeries(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")

		// Series must exist and the data must be able to be unmarshalled
		if _, err := s.Store.GetSeries(sid); err != nil {
			c.AbortWithStatusJSON(404, SeriesReply{Success: false})
			return
		}
		var metadata manga.SeriesMetadata
		if err := c.ShouldBindJSON(&metadata); err != nil {
			log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
			c.AbortWithStatusJSON(400, SeriesReply{Success: false})
			return
		} else if metadata.Title == "" {
			c.AbortWithStatusJSON(400, SeriesReply{Success: false})
			return
		}

		err := s.Store.SetSeriesMetadata(sid, &metadata)
		if err != nil {
			c.AbortWithStatusJSON(500, SeriesReply{Success: false})
			return
		}

		c.JSON(200, SeriesReply{Success: true})
	}
}
