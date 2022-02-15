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
	Success  bool                           `json:"success"`
	Progress map[string]users.EntryProgress `json:"progress"`
}

// EntryProgressRequest for /api/series/:sid/entries/:eid/progress
type EntryProgressRequest struct {
	Progress string `json:"progress"`
}

// EntriesProgressReply for /api/series/:sid/entries/:eid/progress
type EntriesProgressReply struct {
	Success  bool                `json:"success"`
	Progress users.EntryProgress `json:"progress"`
}

// API functions

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

		// Parse and validate the patch request
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

		// Get the series progress
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
			entryExists := sp.HasEntry(e.Hash)
			if !entryExists {
				sp.SetEntry(e.Hash, users.NewEntryProgress(e.Pages, e.Title))
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

		ep, _, _, err := GetEntryProgressInternal(uid, sid, eid, s)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, EntriesProgressReply{Success: true, Progress: ep})
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
		sp.SetEntry(eid, ep)
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

// Internal functions

func GetEntryProgressInternal(uid, sid, eid string, s *server.Server) (users.EntryProgress, users.SeriesProgress, users.CatalogProgress, error) {
	// Ensure series exists
	sp, cp, err := GetSeriesProgressInternal(uid, sid, s)
	if err != nil {
		return users.EntryProgress{}, users.SeriesProgress{}, users.CatalogProgress{}, err
	}

	// Ensure entry exists
	e, err := s.Store.GetEntry(sid, eid)
	if err != nil {
		return users.EntryProgress{}, users.SeriesProgress{}, users.CatalogProgress{}, err
	}

	// Get the entry progress
	ep, err := sp.GetEntry(e.Hash)
	if err != nil || ep.Total != e.Pages {
		// If the entry doesn't exist or if the number
		// of pages has changed then recreate it
		ep = users.NewEntryProgress(e.Pages, e.Title)
		// Update the entry progress in the database
		sp.SetEntry(e.Hash, ep)
		cp.SetSeries(sid, sp)
		err = s.Store.ChangeProgress(uid, cp)
		if err != nil {
			return users.EntryProgress{}, users.SeriesProgress{}, users.CatalogProgress{}, err
		}
	}

	return ep, sp, cp, nil
}

func GetSeriesProgressInternal(uid, sid string, s *server.Server) (users.SeriesProgress, users.CatalogProgress, error) {
	// Ensure the series and its entries exist
	series, err := s.Store.GetSeries(sid)
	if err != nil {
		return users.SeriesProgress{}, users.CatalogProgress{}, err
	}
	entries, err := s.Store.GetEntries(sid)
	if err != nil {
		return users.SeriesProgress{}, users.CatalogProgress{}, err
	}

	// Get the user progress
	progress, err := s.Store.GetUserProgress(uid)
	if err != nil {
		return users.SeriesProgress{}, users.CatalogProgress{}, err
	}

	// Get the series progress
	seriesProgress := progress.GetSeries(sid)

	// If the series is empty or if it's changed we need
	// to recreate the progress
	a := seriesProgress.Empty()
	b := len(seriesProgress.Entries) != len(entries)
	c := false
	if !seriesProgress.Empty() {
		for _, entry := range entries {
			// If the series progress does not contain an
			// entry we have in the current database we flag
			// that we should recreate
			if !seriesProgress.HasEntry(entry.Hash) {
				c = true
			}
		}
	}

	if a || b || c {
		// Keep a record of the old progress
		// so we can copy over old records
		oldP := seriesProgress
		// Create the new progress
		seriesProgress = users.NewSeriesProgress(len(entries), series.Title)
		// Copy over the old progress
		if len(oldP.Entries) > 0 {
			for eid, ep := range oldP.Entries {
				// If the current series has this entry
				// then copy its progress over
				for _, entry := range entries {
					if eid == entry.Hash {
						seriesProgress.SetEntry(eid, ep)
					}
				}
			}
		}
		// Add the new series progress to the catalog
		progress.AddSeries(sid, seriesProgress)
		// Save the newly created progress
		err := s.Store.ChangeProgress(uid, progress)
		if err != nil {
			return users.SeriesProgress{}, users.CatalogProgress{}, err
		}
	}

	return seriesProgress, progress, nil
}
