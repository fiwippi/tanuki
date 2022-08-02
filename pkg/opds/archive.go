package opds

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func GetArchive(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		e, err := s.Store.GetEntry(sid, eid)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		c.FileAttachment(e.Archive.Path, e.Archive.FilenameWithExt())
	}
}
