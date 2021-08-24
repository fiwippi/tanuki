package series

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
)

var ErrNoEntryProgress = errors.New("entry progress does not exist")

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

func GetSeriesProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		uid := c.GetString("uid")

		p, _, err := GetSeriesProgressInternal(uid, sid, s)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// Return the progress
		c.JSON(200, SeriesProgressReply{Success: true, Progress: p.Entries})
	}
}

func PatchSeriesProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		uid := c.GetString("uid")

		var data SeriesProgressRequest
		if err := c.ShouldBindJSON(&data); err != nil {
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
			c.AbortWithError(500, err)
			return
		}

		// Ensure series is filled out with a non-nil entries
		entries, err := s.Store.GetEntries(sid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		for _, e := range entries {
			ep := sp.GetEntryProgress(e.Order - 1)
			if ep == nil {
				ep = users.NewEntryProgress(e.Pages)
				err := sp.SetEntryProgress(e.Order-1, ep)
				if err != nil {
					c.AbortWithError(500, err)
					return
				}
			}

		}

		switch data.Progress {
		case "100%":
			sp.SetAllRead()
		case "0%":
			sp.SetAllUnread()
		}

		// Save the series progress
		cp.SetSeries(sid, sp)
		err = s.Store.ChangeProgress(uid, cp)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// Return the progress
		c.JSON(200, SeriesProgressReply{Success: true})
	}
}

func GetEntryProgress(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		uid := c.GetString("uid")

		p, _, _, err := GetEntryProgressInternal(uid, sid, eid, s)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, EntriesProgressReply{Success: true, Progress: p})
	}
}

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

		ep, sp, cp, err := GetEntryProgressInternal(uid, sid, eid, s)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		if data.Progress == "100%" {
			ep.SetRead()
		} else if data.Progress == "0%" {
			ep.SetUnread()
		} else {
			ep.Set(num)
		}

		// Save the entry progress
		cp.SetSeries(sid, sp)
		err = s.Store.ChangeProgress(uid, cp)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// Return the progress
		c.JSON(200, EntriesProgressReply{Success: true})
	}
}

func GetEntryProgressInternal(uid, sid, eid string, s *server.Server) (*users.EntryProgress, *users.SeriesProgress, *users.CatalogProgress, error) {
	sp, cp, err := GetSeriesProgressInternal(uid, sid, s)
	if err != nil {
		return nil, nil, nil, err
	}

	e, err := s.Store.GetEntry(sid, eid)
	if err != nil {
		return nil, nil, nil, err
	}

	// Entry must be within the size of the series progress
	index := e.Order - 1
	if index >= len(sp.Entries) || index < 0 {
		return nil, nil, nil, ErrNoEntryProgress
	}

	// Get the entry progress
	ep := sp.GetEntryProgress(e.Order - 1)
	if ep == nil {
		// If it doesn't exist then create it
		ep = users.NewEntryProgress(e.Pages)
		err := sp.SetEntryProgress(e.Order-1, ep)
		if err != nil {
			return nil, nil, nil, err
		}

		// Save the progress to the db
		cp.SetSeries(sid, sp)
		err = s.Store.ChangeProgress(uid, cp)
		if err != nil {
			return nil, nil, nil, err
		}

	}

	return ep, sp, cp, nil
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
	// If the series exists but the progress for it doesnt
	// exist then create the new progress for the user, or
	// if the number of entries has changed
	if seriesProgress == nil || len(seriesProgress.Entries) != len(entries) {
		// Keep track of the old progress
		oldProgress := seriesProgress
		// Create the new series progress
		progress.AddSeries(sid, len(entries))
		seriesProgress = progress.GetSeries(sid)
		// Copy over the old progress
		if oldProgress != nil && len(oldProgress.Entries) > 0 {
			for i, e := range oldProgress.Entries {
				err := seriesProgress.SetEntryProgress(i, e)
				if err != nil {
					return nil, nil, err
				}
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
