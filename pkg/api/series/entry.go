package series

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"strconv"
)

// TODO remove the structs which aren't needed if we don't need to unmarshal data from them in the frontent

// SeriesEntryReply for /api/series/:id/entries/:eid
type SeriesEntryReply struct {
	Success bool      `json:"success"`
	Data    api.Entry `json:"data"`
}

// PATCH /api/series/:sid/entries/:eid
func PatchEntry(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		// Series must exist and the data must be able to be unmarshalled
		if _, err := s.Store.GetEntry(sid, eid); err != nil {
			c.AbortWithStatusJSON(404, SeriesEntryReply{Success: false})
			return
		}
		var metadata manga.EntryMetadata
		if err := c.ShouldBindJSON(&metadata); err != nil {
			log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
			c.AbortWithStatusJSON(400, SeriesEntryReply{Success: false})
			return
		} else if metadata.Title == "" {
			c.AbortWithStatusJSON(400, SeriesEntryReply{Success: false})
			return
		}

		err := s.Store.SetEntryMetadata(sid, eid, &metadata)
		if err != nil {
			c.AbortWithStatusJSON(500, SeriesEntryReply{Success: false})
			return
		}

		c.JSON(200, SeriesEntryReply{Success: true})
	}
}

// GET /api/series/:sid/entries/:eid
func GetEntry(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		e, err := s.Store.GetEntry(sid, eid)
		if err != nil {
			c.AbortWithStatusJSON(500, SeriesEntryReply{Success: false})
			return
		}
		c.JSON(200, SeriesEntryReply{Success: true, Data: *e})
	}
}

// GET /api/series/:sid/entries/:eid/archive
func GetEntryArchive(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		a, err := s.Store.GetEntryArchive(sid, eid)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		c.FileAttachment(a.Path, a.FilenameWithExt())
	}
}

// GET /api/series/:sid/entries/:eid/page/:num
func GetEntryPage(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		numStr := c.Param("num")

		num, err := strconv.Atoi(numStr)
		if err != nil {
			c.AbortWithStatus(400)
			return
		}

		a, err := s.Store.GetEntryArchive(sid, eid)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		p, err := s.Store.GetEntryPage(sid, eid, num)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		r, size, err := a.ReaderForFile(p.Path)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		c.DataFromReader(200, size, p.ImageType.MimeType(), r, nil)
	}
}
