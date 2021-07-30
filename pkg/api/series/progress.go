package series

import (
	"errors"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"strconv"
)

// Progress can be defined as 100%, 0% or an int
// representing the page number the user is on,
// page numbers can only be used when setting progress
// for entries, progress for series must be 0% or 100%

// SeriesProgressRequest for /api/series/:sid/progress
type SeriesProgressRequest struct {
	Progress string `json:"progress"`
}

// SeriesProgressReply for /api/series/:sid/progress
type SeriesProgressReply struct {
	Success  bool                   `json:"success"`
	Progress []*users.EntryProgress `json:"progress"`
}

// EntryProgressRequest for /api/series/:sid/entries/:eid/progress
type EntryProgressRequest struct {
	Progress string `json:"progress"`
}

// EntriesProgressReply for /api/series/:sid/entries/:eid/progress
type EntriesProgressReply struct {
	Success  bool                 `json:"success"`
	Progress *users.EntryProgress `json:"progress"`
}

// GET /api/series/:sid/progress
func GetSeriesProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		uid := c.GetString("uid")

		p, _, err := GetSeriesProgressInternal(uid, sid, s)
		if err != nil {
			log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
			c.AbortWithStatusJSON(500, SeriesProgressReply{Success: false})
			return
		}

		// Return the progress
		c.JSON(200, SeriesProgressReply{Success: true, Progress: p.Entries})
	}
}

// GET /api/series/:sid/progress
func PatchSeriesProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		uid := c.GetString("uid")

		var data SeriesProgressRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("could not bind json")
			c.AbortWithStatusJSON(400, SeriesProgressReply{Success: false})
			return
		}
		if data.Progress != "0%" && data.Progress != "100%" {
			log.Debug().Err(errors.New("series progress is not specified as 0% or 100%")).Str("progress", data.Progress).Msg("")
			c.AbortWithStatusJSON(400, SeriesProgressReply{Success: false})
			return
		}

		sp, cp, err := GetSeriesProgressInternal(uid, sid, s)
		if err != nil {
			log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
			c.AbortWithStatusJSON(500, SeriesProgressReply{Success: false})
			return
		}

		switch data.Progress {
		case "100%":
			sp.SetAllRead()
		case "0%":
			sp.SetAllUnread()
		}

		// Save the series progress
		err = s.Store.ChangeProgress(uid, cp)
		if err != nil {
			c.AbortWithStatusJSON(500, SeriesProgressReply{Success: false})
			return
		}

		// Return the progress
		c.JSON(200, SeriesProgressReply{Success: true})
	}
}

// GET /api/series/:sid/entries/:eid/progress
func GetEntryProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		uid := c.GetString("uid")

		p, _, err := GetSeriesProgressInternal(uid, sid, s)
		if err != nil {
			log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
			c.AbortWithStatusJSON(500, EntriesProgressReply{Success: false})
			return
		}

		o, err := s.Store.GetEntryOrder(sid, eid)
		if err != nil {
			log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not get entry order")
			c.AbortWithStatusJSON(500, EntriesProgressReply{Success: false})
			return
		}

		c.JSON(200, EntriesProgressReply{Success: true, Progress: p.GetEntryProgress(o - 1)})
	}
}

// GET /api/series/:sid/progress
func PatchEntryProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		uid := c.GetString("uid")

		var data EntryProgressRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("could not bind json")
			c.AbortWithStatusJSON(400, EntriesProgressReply{Success: false})
			return
		}
		num, err := strconv.Atoi(data.Progress)
		if data.Progress != "0%" && data.Progress != "100%" && err != nil {
			log.Debug().Err(errors.New("invalid entry progress")).Str("progress", data.Progress).Msg("")
			c.AbortWithStatusJSON(400, EntriesProgressReply{Success: false})
			return
		}

		sp, cp, err := GetSeriesProgressInternal(uid, sid, s)
		if err != nil {
			log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
			c.AbortWithStatusJSON(500, EntriesProgressReply{Success: false})
			return
		}

		o, err := s.Store.GetEntryOrder(sid, eid)
		if err != nil {
			log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not get entry order")
			c.AbortWithStatusJSON(500, EntriesProgressReply{Success: false})
			return
		}

		ep := sp.GetEntryProgress(o - 1)
		if data.Progress == "100%" {
			ep.SetRead()
		} else if data.Progress == "0%" {
			ep.SetUnread()
		} else {
			ep.Set(num)
		}

		// Save the entry progress
		err = s.Store.ChangeProgress(uid, cp)
		if err != nil {
			c.AbortWithStatusJSON(500, EntriesProgressReply{Success: false})
			return
		}

		// Return the progress
		c.JSON(200, EntriesProgressReply{Success: true})
	}
}

func GetEntryProgressInternal(uid, sid, eid string, s *server.Server) (*users.EntryProgress, error) {
	p, _, err := GetSeriesProgressInternal(uid, sid, s)
	if err != nil {
		log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
		return nil, err
	}

	o, err := s.Store.GetEntryOrder(sid, eid)
	if err != nil {
		log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not get entry order")
		return nil, err
	}

	return p.GetEntryProgress(o - 1), nil
}

func GetSeriesProgressInternal(uid, sid string, s *server.Server) (*users.SeriesProgress, *users.CatalogProgress, error) {
	// Ensure the series and its entries exist
	entries, err := s.Store.GetEntries(sid)
	if err != nil {
		return nil, nil, err
	}

	// Get the user progress
	progress, err := s.Store.GetUserProgress(uid)
	if err != nil {
		return nil, nil, err
	}

	// Get the series progress
	seriesProgress := progress.GetSeries(sid)
	if seriesProgress == nil {
		// If the series exists but the progress for it doesnt
		// exist then create the new progress for the user
		progress.AddSeries(sid, len(entries))
		seriesProgress = progress.GetSeries(sid)
		for i, e := range entries {
			err := seriesProgress.SetEntryProgress(i, users.NewEntryProgress(e.Pages))
			if err != nil {
				return nil, nil, err
			}
		}

		// Save the newly created progress
		err := s.Store.ChangeProgress(uid, progress)
		if err != nil {
			return nil, nil, err
		}
	}

	return seriesProgress, progress, err
}
