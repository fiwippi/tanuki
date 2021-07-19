package tanuki

import (
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
func apiGetSeriesList(c *gin.Context) {
	list := db.GenerateSeriesList()
	c.JSON(200, api.SeriesListReply{Success: true, List: list})
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
	var data api.Series
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.SeriesReply{Success: false})
		return
	} else if data.Title == "" {
		c.AbortWithStatusJSON(400, api.SeriesReply{Success: false})
		return
	}

	m := core.NewSeriesMetadata()
	m.Title = data.Title
	m.Author = data.Author
	m.DateReleased = data.DateReleased

	err := db.SetSeriesMetadata(id, m)
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
	var data api.SeriesEntry
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.SeriesEntryReply{Success: false})
		return
	} else if data.Title == "" {
		c.AbortWithStatusJSON(400, api.SeriesEntryReply{Success: false})
		return
	}

	m := core.NewEntryMetadata()
	m.Title = data.Title
	m.Author = data.Author
	m.DateReleased = data.DateReleased
	m.Chapter = data.Chapter
	m.Volume = data.Volume

	err := db.SetEntryMetadata(sid, eid, m)
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
		img,  mimetype, err = db.GetSeriesCoverBytes(id)
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

// GET /api/series/:sid
func apiGetSeriesEntries(c *gin.Context) {
	id := c.Param("sid")
	entries, err := db.GetSeriesEntries(id)
	if err != nil {
		c.AbortWithStatusJSON(500, api.SeriesEntriesReply{Success: false})
		return
	}
	c.JSON(200, api.SeriesEntriesReply{Success: true, List: entries})
}

// GET /api/series/:sid/entries/:eid
func apiGetSeriesEntry(c *gin.Context) {
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
func apiGetSeriesEntryCover(c *gin.Context) {
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
		img, mimetype, err = db.GetSeriesEntryCoverBytes(sid, eid)
	}

	if err != nil || len(img) == 0 {
		log.Debug().Err(err).Int("img length", len(img)).Str("sid", sid).Str("eid", eid).Msg("failed to get entry")
		c.AbortWithStatus(500)
		return
	}
	c.Data(200, mimetype, img)
}

// GET /api/series/:sid/entries/:eid/archive
func apiGetSeriesEntryArchive(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")

	a, err := db.GetSeriesEntryArchive(sid, eid)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.FileAttachment(a.Path, a.FilenameWithExt())
}

// GET /api/series/:sid/entries/:eid/page/:num
func apiGetSeriesEntryPage(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")
	numStr := c.Param("num")

	num, err := strconv.Atoi(numStr)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	a, err := db.GetSeriesEntryArchive(sid, eid)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	p, err := db.GetSeriesEntryPage(sid, eid, num)
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