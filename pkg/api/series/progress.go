package series

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

// Progress can be defined as 100%, 0% or an int
// representing the page number the user is on,
// page numbers can only be used when setting progress
// for entries, progress for series must be 0% or 100%

func GetSeriesProgress(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		uid := c.GetString("uid")

		p, err := s.Store.GetSeriesProgress(sid, uid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// Return the progress
		c.JSON(200, gin.H{"progress": p})
	}
}

func PatchSeriesProgress(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		uid := c.GetString("uid")

		data := struct {
			Progress string `json:"progress"`
		}{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithError(400, err)
			return
		}
		if data.Progress != "0%" && data.Progress != "100%" {
			c.AbortWithError(400, fmt.Errorf("progress is not '0%%' or '100%%'"))
			return
		}

		var err error
		switch data.Progress {
		case "100%":
			err = s.Store.SetSeriesProgressRead(sid, uid)
		case "0%":
			err = s.Store.SetSeriesProgressUnread(sid, uid)
		}
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Status(200)
	}
}

func GetEntryProgress(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		uid := c.GetString("uid")

		ep, err := s.Store.GetEntryProgress(sid, eid, uid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, gin.H{"progress": ep})
	}
}

func PatchEntryProgress(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		uid := c.GetString("uid")

		data := struct {
			Progress string `json:"progress"`
		}{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithError(400, err)
			return
		}
		num, err := strconv.Atoi(data.Progress)
		if data.Progress != "0%" && data.Progress != "100%" && err != nil {
			c.AbortWithError(400, fmt.Errorf("invalid entry progress"))
			return
		}

		if data.Progress == "100%" {
			err = s.Store.SetEntryProgressRead(sid, eid, uid)
		} else if data.Progress == "0%" {
			err = s.Store.SetEntryProgressUnread(sid, eid, uid)
		} else {
			err = s.Store.SetEntryProgressAmount(sid, eid, uid, num)
		}

		c.Status(200)
	}
}
