package tanuki

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/core"
)

// GET /api/series
func apiGetCatalog(c *gin.Context) {
	list := db.GetCatalog()
	c.JSON(200, api.CatalogReply{Success: true, List: list})
}

// GET /api/series/:sid
func apiGetSeries(c *gin.Context) {
	id := c.Param("sid")
	s, err := db.GetSeries(id)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesReply{Success: false})
		return
	}
	c.JSON(200, api.SeriesReply{Success: true, Data: *s})
}

// PATCH /api/series/:sid
func apiPatchSeries(c *gin.Context) {
	id := c.Param("sid")

	// Series must exist and the data must be able to be unmarshalled
	if _, err := db.GetSeries(id); err != nil {
		c.AbortWithStatusJSON(404, api.SeriesReply{Success: false})
		return
	}
	var metadata api.EditableSeriesMetadata
	if err := c.ShouldBindJSON(&metadata); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.SeriesReply{Success: false})
		return
	} else if metadata.Title == "" {
		c.AbortWithStatusJSON(400, api.SeriesReply{Success: false})
		return
	}

	err := db.SetSeriesMetadata(id, &metadata)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesReply{Success: false})
		return
	}

	c.JSON(200, api.SeriesReply{Success: true})
}

// PATCH /api/series/:sid/entries/:eid
func apiPatchEntry(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")

	// Series must exist and the data must be able to be unmarshalled
	if _, err := db.GetEntry(sid, eid); err != nil {
		c.AbortWithStatusJSON(404, api.SeriesEntryReply{Success: false})
		return
	}
	var metadata api.EditableEntryMetadata
	if err := c.ShouldBindJSON(&metadata); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.SeriesEntryReply{Success: false})
		return
	} else if metadata.Title == "" {
		c.AbortWithStatusJSON(400, api.SeriesEntryReply{Success: false})
		return
	}

	err := db.SetEntryMetadata(sid, eid, &metadata)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesEntryReply{Success: false})
		return
	}

	c.JSON(200, api.SeriesEntryReply{Success: true})
}

// GET /api/series/:sid/cover?thumbnail={true,false}
func apiGetSeriesCover(c *gin.Context) {
	id := c.Param("sid")
	thumbnail := c.DefaultQuery("thumbnail", "false")

	var img []byte
	var err error
	var mimetype string
	if thumbnail == "true" {
		img, mimetype, err = db.GetSeriesThumbnail(id)
	} else {
		img, mimetype, err = db.GetSeriesCoverFile(id)
	}

	if err != nil {
		log.Debug().Err(err).Str("sid", id).Msg("failed to get cover")
		c.AbortWithStatus(500)
		return
	}

	c.Data(200, mimetype, img)
}

// PATCH /api/series/:sid/cover
func apiPatchSeriesCover(c *gin.Context) {
	id := c.Param("sid")

	// Ensure the series exists
	_, err := db.GetSeries(id)
	if err != nil {
		c.AbortWithStatusJSON(404, api.PatchCoverReply{Success: false})
		return
	}

	// Validate form data
	filename := c.PostForm("filename")
	file, err := c.FormFile("file")
	if err != nil || filename == "" {
		log.Debug().Err(err).Str("filename", filename).Msg("invalid form for patching cover")
		c.AbortWithStatusJSON(400, api.PatchCoverReply{Success: false})
		return
	}

	// Save the file
	t, err := db.GetSeriesFolderTitle(id)
	if err != nil {
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	imageType, err := core.GetImageType(filepath.Ext(filename))
	if err != nil {
		log.Debug().Err(err).Str("filename", filepath.Ext(filename)).Msg("failed getting image type for new cover")
		c.AbortWithStatusJSON(400, api.PatchCoverReply{Success: false})
		return
	}

	fp := fmt.Sprintf("%s/%s/.tanuki/%s.%s", conf.Paths.Library, t, id, imageType.String())
	err = fse.EnsureFileDir(fp)
	if err != nil {
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	err = c.SaveUploadedFile(file, fp)
	if err != nil {
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Set the new series cover data
	cover, err := db.GetSeriesCover(id)
	if err != nil {
		log.Debug().Err(err).Str("sid", id).Msg("could not get series cover")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}
	cover.Fp = fp
	cover.ImageType = imageType
	err = db.SetSeriesCover(id, cover)
	if err != nil {
		log.Debug().Err(err).Str("sid", id).Msg("could not save series cover")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Generate the new thumbnail
	if err := db.GenerateSeriesThumbnail(id, true); err != nil {
		log.Debug().Err(err).Str("sid", id).Msg("could not create thumbnail for new cover")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Reply to user
	c.JSON(200, api.PatchCoverReply{Success: true})
}

// PATCH /api/series/:sid/entries/:eid/cover
func apiPatchEntryCover(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")

	// Ensure the series exists
	_, err := db.GetEntry(sid, eid)
	if err != nil {
		c.AbortWithStatusJSON(404, api.PatchCoverReply{Success: false})
		return
	}

	// Validate form data
	filename := c.PostForm("filename")
	file, err := c.FormFile("file")
	if err != nil || filename == "" {
		log.Debug().Err(err).Str("filename", filename).Msg("invalid form for patching cover")
		c.AbortWithStatusJSON(400, api.PatchCoverReply{Success: false})
		return
	}

	// Save the file
	t, err := db.GetSeriesFolderTitle(sid)
	if err != nil {
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}
	imageType, err := core.GetImageType(filepath.Ext(filename))
	if err != nil {
		log.Debug().Err(err).Str("filename", filepath.Ext(filename)).Msg("failed getting image type for new cover")
		c.AbortWithStatusJSON(400, api.PatchCoverReply{Success: false})
		return
	}
	fp := fmt.Sprintf("%s/%s/.tanuki/%s.%s", conf.Paths.Library, t, eid, imageType.String())
	err = fse.EnsureFileDir(fp)
	if err != nil {
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}
	err = c.SaveUploadedFile(file, fp)
	if err != nil {
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Set the new entry cover data
	cover, err := db.GetEntryCover(sid, eid)
	if err != nil {
		log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not get cover")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}
	cover.Fp = fp
	cover.ImageType = imageType
	err = db.SetEntryCover(sid, eid, cover)
	if err != nil {
		log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not save series cover")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Generate the new thumbnail
	if err := db.GenerateEntryThumbnail(sid, eid, true); err != nil {
		log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not create thumbnail for new cover")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Reply to user
	c.JSON(200, api.PatchCoverReply{Success: true})
}

// DELETE /api/series/:sid/entries/:eid/cover
func apiDeleteEntryCover(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")

	// Ensure the series exists
	_, err := db.GetEntry(sid, eid)
	if err != nil {
		c.AbortWithStatusJSON(404, api.PatchCoverReply{Success: false})
		return
	}

	err = db.DeleteEntryCover(sid, eid)
	if err != nil {
		log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not delete cover and thumbnail")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Reply to user
	c.JSON(200, api.PatchCoverReply{Success: true})
}

// DELETE /api/series/:sid/cover
func apiDeleteSeriesCover(c *gin.Context) {
	id := c.Param("sid")

	// Ensure the series exists
	_, err := db.GetSeries(id)
	if err != nil {
		c.AbortWithStatusJSON(404, api.PatchCoverReply{Success: false})
		return
	}

	err = db.DeleteSeriesCover(id)
	if err != nil {
		log.Debug().Err(err).Str("sid", id).Msg("could not delete series cover and thumbnail")
		c.AbortWithStatusJSON(500, api.PatchCoverReply{Success: false})
		return
	}

	// Reply to user
	c.JSON(200, api.PatchCoverReply{Success: true})
}

// GET /api/series/:sid/entries
func apiGetSeriesEntries(c *gin.Context) {
	sid := c.Param("sid")
	entries, err := db.GetEntries(sid)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesEntriesReply{Success: false})
		return
	}
	c.JSON(200, api.SeriesEntriesReply{Success: true, List: entries})
}

// GET /api/series/:sid/entries/:eid
func apiGetEntry(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")
	e, err := db.GetEntry(sid, eid)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesEntryReply{Success: false})
		return
	}
	c.JSON(200, api.SeriesEntryReply{Success: true, Data: *e})
}

// GET /api/series/:sid/entries/:eid/cover?thumbnail={true,false}
func apiGetEntryCover(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")
	thumbnail := c.DefaultQuery("thumbnail", "false")

	var img []byte
	var err error
	var mimetype string
	if thumbnail == "true" {

		img, mimetype, err = db.GetEntryThumbnail(sid, eid)

		// If thumbnail doesn't exist try and recreate it
		if len(img) == 0 {
			err = db.GenerateEntryThumbnail(sid, eid, true)
			if err != nil {
				c.AbortWithStatus(500)
				return
			}

			img, mimetype, err = db.GetEntryThumbnail(sid, eid)
		}
	} else {
		img, mimetype, err = db.GetEntryCoverFile(sid, eid)
	}

	if err != nil || len(img) == 0 {
		log.Debug().Err(err).Int("img length", len(img)).Str("sid", sid).Str("eid", eid).Msg("failed to get entry")
		c.AbortWithStatus(500)
		return
	}
	c.Data(200, mimetype, img)
}

// GET /api/series/:sid/entries/:eid/archive
func apiGetEntryArchive(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")

	a, err := db.GetEntryArchive(sid, eid)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.FileAttachment(a.Path, a.FilenameWithExt())
}

// GET /api/series/:sid/entries/:eid/page/:num
func apiGetEntryPage(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")
	numStr := c.Param("num")

	num, err := strconv.Atoi(numStr)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	a, err := db.GetEntryArchive(sid, eid)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	p, err := db.GetEntryPage(sid, eid, num)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	r, size, err := a.FileReader(p.Path)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.DataFromReader(200, size, p.ImageType.MimeType(), r, nil)
}

// GET /api/series/:sid/tags
func apiGetSeriesTags(c *gin.Context) {
	id := c.Param("sid")
	t, err := db.GetSeriesTags(id)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesTagsReply{Success: false})
		return
	}
	c.JSON(200, api.SeriesTagsReply{Success: true, Tags: t.List()})
}

// PATCH /api/series/:sid/tags
func apiPatchSeriesTags(c *gin.Context) {
	var data api.SeriesTagsRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.SeriesTagsReply{Success: false})
		return
	}

	id := c.Param("sid")
	if err := db.SetSeriesTags(id, data.Tags); err != nil {
		log.Debug().Err(err).Str("series", id).Msg("failed to set tags")
		c.AbortWithStatusJSON(500, api.SeriesTagsReply{Success: false})
		return
	}

	c.JSON(200, api.SeriesTagsReply{Success: true})
}

// GET /api/catalog/progress
func apiGetCatalogProgress(c *gin.Context) {
	uid := c.GetString("uid")

	user, err := db.GetUser(uid)
	if err != nil {
		log.Debug().Err(err).Str("uid", uid).Msg("could not get user")
		c.AbortWithStatusJSON(500, api.CatalogProgressReply{Success: false})
		return
	}

	// Return the progress
	c.JSON(200, api.CatalogProgressReply{Success: true, Progress: user.Progress.Data})
}

// GET /api/series/:sid/progress
func apiGetSeriesProgress(c *gin.Context) {
	sid := c.Param("sid")
	uid := c.GetString("uid")

	p, _, err := getSeriesProgress(uid, sid)
	if err != nil {
		log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
		c.AbortWithStatusJSON(500, api.SeriesProgressReply{Success: false})
		return
	}

	// Return the progress
	c.JSON(200, api.SeriesProgressReply{Success: true, Progress: p.Entries})
}

// Progress can be defined as 100%, 0% or an int
// representing the page number the user is on,
// page numbers can only be used when setting progress
// for entries, progress for series must be 0% or 100%
// GET /api/series/:sid/progress
func apiPatchSeriesProgress(c *gin.Context) {
	sid := c.Param("sid")
	uid := c.GetString("uid")

	var data api.SeriesProgressRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("could not bind json")
		c.AbortWithStatusJSON(400, api.SeriesProgressReply{Success: false})
		return
	}
	if data.Progress != "0%" && data.Progress != "100%" {
		log.Debug().Err(errors.New("series progress is not specified as 0% or 100%")).Str("progress", data.Progress).Msg("")
		c.AbortWithStatusJSON(400, api.SeriesProgressReply{Success: false})
		return
	}

	sp, cp, err := getSeriesProgress(uid, sid)
	if err != nil {
		log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
		c.AbortWithStatusJSON(500, api.SeriesProgressReply{Success: false})
		return
	}

	switch data.Progress {
	case "100%":
		sp.SetAllRead()
	case "0%":
		sp.SetAllUnread()
	}

	// Save the series progress
	err = db.ChangeProgress(uid, cp)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesProgressReply{Success: false})
		return
	}

	// Return the progress
	c.JSON(200, api.SeriesProgressReply{Success: true})
}

// GET /api/series/:sid/entries/:eid/progress
func apiGetEntryProgress(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")
	uid := c.GetString("uid")

	p, _, err := getSeriesProgress(uid, sid)
	if err != nil {
		log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
		c.AbortWithStatusJSON(500, api.EntriesProgressReply{Success: false})
		return
	}

	o, err := db.GetEntryOrder(sid, eid)
	if err != nil {
		log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not get entry order")
		c.AbortWithStatusJSON(500, api.EntriesProgressReply{Success: false})
		return
	}

	c.JSON(200, api.EntriesProgressReply{Success: true, Progress: p.GetEntryProgress(o - 1)})
}

// GET /api/series/:sid/progress
func apiPatchEntryProgress(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")
	uid := c.GetString("uid")

	var data api.EntryProgressRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("could not bind json")
		c.AbortWithStatusJSON(400, api.EntriesProgressReply{Success: false})
		return
	}
	num, err := strconv.Atoi(data.Progress)
	if data.Progress != "0%" && data.Progress != "100%" && err != nil {
		log.Debug().Err(errors.New("invalid entry progress")).Str("progress", data.Progress).Msg("")
		c.AbortWithStatusJSON(400, api.EntriesProgressReply{Success: false})
		return
	}

	sp, cp, err := getSeriesProgress(uid, sid)
	if err != nil {
		log.Debug().Err(err).Str("uid", uid).Str("sid", sid).Msg("could not get progress")
		c.AbortWithStatusJSON(500, api.EntriesProgressReply{Success: false})
		return
	}

	o, err := db.GetEntryOrder(sid, eid)
	if err != nil {
		log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not get entry order")
		c.AbortWithStatusJSON(500, api.EntriesProgressReply{Success: false})
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
	err = db.ChangeProgress(uid, cp)
	if err != nil {
		c.AbortWithStatusJSON(500, api.EntriesProgressReply{Success: false})
		return
	}

	// Return the progress
	c.JSON(200, api.EntriesProgressReply{Success: true})
}

func getSeriesProgress(uid, sid string) (*core.SeriesProgress, *core.CatalogProgress, error) {
	// Get the user
	user, err := db.GetUser(uid)
	if err != nil {
		return nil, nil, err
	}

	// Ensure the series and its entries exist
	entries, err := db.GetEntries(sid)
	if err != nil {
		return nil, nil, err
	}

	// Get the progress for the series
	p := user.Progress.GetSeries(sid)
	if p == nil {
		// If the series exists but the progress for it doesnt
		// exist then create the new progress for the user
		user.Progress.AddSeries(sid, len(entries))
		p = user.Progress.GetSeries(sid)
		for i, e := range entries {
			err := p.SetEntryProgress(i, core.NewEntryProgress(e.Pages))
			if err != nil {
				return nil, nil, err
			}
		}

		// Save the newly created progress
		err := db.ChangeProgress(uid, user.Progress)
		if err != nil {
			return nil, nil, err
		}
	}

	return p, user.Progress, err
}
