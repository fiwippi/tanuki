package tanuki

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/api"
)

// These are api routes each user can use to get and
// modify properties about themselves, users are
// expected to provide the cookie identifying themselves
// in order to access/edit their own data unless they're
// an admin

// GET /api/user/:property
func apiGetUserProperty(c *gin.Context) {
	uid := c.GetString("uid")
	u, err := db.GetUser(uid)
	if err != nil {
		c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
		return
	}

	switch c.Param("property") {
	case "type":
		c.JSON(200, api.UserPropertyReply{Success: true, Type: u.Type})
	case "name":
		c.JSON(200, api.UserPropertyReply{Success: true, Username: u.Name})
	case "progress":
		series := c.DefaultQuery("series", "")
		entry := c.DefaultQuery("entry", "")

		// Can't specify entries without series
		if entry != "" && series == "" {
			c.AbortWithStatusJSON(400, api.UserPropertyReply{Success: false})
			return
		}
		// Ensure series (and entries) exist
		if !db.HasSeries(series) {
			log.Debug().Err(err).Str("sid", series).Msg("couldn't find series")
			c.AbortWithStatusJSON(404, api.UserPropertyReply{Success: false})
			return
		}
		if entry != "" && !db.HasSeriesEntry(series, entry) {
			log.Debug().Err(err).Str("sid", series).Str("eid", entry).Msg("couldn't find entry")
			c.AbortWithStatusJSON(404, api.UserPropertyReply{Success: false})
			return
		}

		// If they exist and the progress doesn't exist for the series then we need
		// to create a new progress for it
		save := false
		switch entry == "" {
		case true:
			if !u.ProgressTracker.HasSeries(series) {
				u.ProgressTracker.AddSeries(series)
				save = true
			}
		case false:
			if !u.ProgressTracker.HasEntry(series, entry) {
				e, err := db.GetEntry(series, entry)
				if err != nil {
					c.AbortWithStatusJSON(400, api.UserPropertyReply{Success: false})
					return
				}
				u.ProgressTracker.AddEntry(series, entry, e.Pages)
				save = true
			}
		}

		// Save the created progress back to the user
		if save {
			err := db.ChangeProgressTracker(uid, u.ProgressTracker)
			if err != nil {
				c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
				return
			}
		}

		// Send the progress back to the user
		if entry == "" {
			sp := u.ProgressTracker.ProgressSeries(series)
			c.JSON(200, api.UserPropertyReply{Success: true, ProgressPercent: sp.Percent(), ProgressPage: sp.Current})
		} else {
			ep := u.ProgressTracker.ProgressEntry(series, entry)
			c.JSON(200, api.UserPropertyReply{Success: true, ProgressPercent: ep.Percent(), ProgressPage: ep.Current})
		}
	default:
		c.AbortWithStatusJSON(400, api.UserPropertyReply{Success: false})
	}
}

// Progress can be defined as 100%, 0% or an int
// representing the page number the user is on,
// page numbers can only be used when setting progress
// for entries, progress for series must be 0% or 100%
// PATCH /api/user/progress
func apiPatchUserProgress(c *gin.Context) {
	series := c.DefaultQuery("series", "")
	entry := c.DefaultQuery("entry", "")

	// Can't specify entries without series
	if entry != "" && series == "" {
		c.AbortWithStatusJSON(400, api.UserPropertyReply{Success: false})
		return
	}

	// Ensure user exists
	uid := c.GetString("uid")
	_, err := db.GetUser(uid)
	if err != nil {
		log.Debug().Err(err).Msg("failed to verify user")
		c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
		return
	}

	// Ensure series (and entries) exist
	if !db.HasSeries(series) {
		c.AbortWithStatusJSON(404, api.UserPropertyReply{Success: false})
		return
	}
	if entry != "" && !db.HasSeriesEntry(series, entry) {
		c.AbortWithStatusJSON(404, api.UserPropertyReply{Success: false})
		return
	}

	// Ensure valid data supplied
	var data api.UserProgressRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.UserPropertyReply{Success: false})
		return
	}
	num, err := strconv.Atoi(data.Progress)
	if data.Progress != "0%" && data.Progress != "100%" && err != nil {
		c.AbortWithStatusJSON(400, api.UserPropertyReply{Success: false})
		return
	}

	switch entry == "" {
	case true:
		// Set series as either AllRead or AllUnread
		if data.Progress == "100%" {
			if err := db.SetSeriesProgressAllRead(uid, series); err != nil {
				log.Debug().Err(err).Msg("failed to set progress to all read")
				c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
				return
			}
		} else if data.Progress == "0%" {
			if err := db.SetSeriesProgressAllUnread(uid, series); err != nil {
				log.Debug().Err(err).Msg("failed to set progress to all unread")
				c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
				return
			}
		} else {
			c.AbortWithStatusJSON(400, api.UserPropertyReply{Success: false})
			return
		}
	case false:
		if data.Progress == "100%" {
			if err := db.SetEntryProgressRead(uid, series, entry); err != nil {
				log.Debug().Err(err).Msg("failed to set progress to read")
				c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
				return
			}
		} else if data.Progress == "0%" {
			if err := db.SetEntryProgressUnread(uid, series, entry); err != nil {
				log.Debug().Err(err).Msg("failed to set progress to unread")
				c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
				return
			}
		} else {
			if err := db.SetSeriesEntryProgressNum(uid, series, entry, num); err != nil {
				log.Debug().Err(err).Int("num", num).Msg("failed to set progress")
				c.AbortWithStatusJSON(500, api.UserPropertyReply{Success: false})
				return
			}
		}
	}

	c.JSON(200, api.UserPropertyReply{Success: true})
}
