package series

import (
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/pkg/server"
)

// PatchCoverReply for /api/series/:sid/cover
type PatchCoverReply struct {
	Success bool `json:"success"`
}

func GetSeriesCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")
		thumbnail := c.DefaultQuery("thumbnail", "false")

		var img []byte
		var err error
		var mimetype string
		if thumbnail == "true" {
			img, mimetype, err = s.Store.GetSeriesThumbnail(id)
		} else {
			img, mimetype, err = s.Store.GetSeriesCoverFile(id)
		}

		if err != nil {
			log.Debug().Err(err).Str("sid", id).Msg("failed to get cover")
			c.AbortWithStatus(500)
			return
		}

		c.Data(200, mimetype, img)
	}
}

func PatchSeriesCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")

		// Ensure the series exists
		_, err := s.Store.GetSeries(id)
		if err != nil {
			c.AbortWithStatusJSON(404, PatchCoverReply{Success: false})
			return
		}

		// Validate form data
		filename := c.PostForm("filename")
		file, err := c.FormFile("file")
		if err != nil || filename == "" {
			log.Debug().Err(err).Str("filename", filename).Msg("invalid form for patching cover")
			c.AbortWithStatusJSON(400, PatchCoverReply{Success: false})
			return
		}

		// Save the file
		t, err := s.Store.GetSeriesFolderTitle(id)
		if err != nil {
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		imageType, err := image.InferType(filename)
		if err != nil {
			log.Debug().Err(err).Str("filename", filepath.Ext(filename)).Msg("failed getting image type for new cover")
			c.AbortWithStatusJSON(400, PatchCoverReply{Success: false})
			return
		}

		fp := fmt.Sprintf("%s/%s/.tanuki/%s.%s", s.Conf.Paths.Library, t, id, imageType.String())
		err = fse.EnsureFileDir(fp)
		if err != nil {
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		err = c.SaveUploadedFile(file, fp)
		if err != nil {
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Set the new series cover data
		cover, err := s.Store.GetSeriesCover(id)
		if err != nil {
			log.Debug().Err(err).Str("sid", id).Msg("could not get series cover")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}
		cover.Fp = fp
		cover.ImageType = imageType
		err = s.Store.SetSeriesCover(id, cover)
		if err != nil {
			log.Debug().Err(err).Str("sid", id).Msg("could not save series cover")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Generate the new thumbnail
		if err := s.Store.GenerateSeriesThumbnail(id, true); err != nil {
			log.Debug().Err(err).Str("sid", id).Msg("could not create thumbnail for new cover")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Reply to user
		c.JSON(200, PatchCoverReply{Success: true})
	}
}

func DeleteSeriesCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("sid")

		// Ensure the series exists
		_, err := s.Store.GetSeries(id)
		if err != nil {
			c.AbortWithStatusJSON(404, PatchCoverReply{Success: false})
			return
		}

		err = s.Store.DeleteSeriesCover(id)
		if err != nil {
			log.Debug().Err(err).Str("sid", id).Msg("could not delete series cover and thumbnail")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Reply to user
		c.JSON(200, PatchCoverReply{Success: true})
	}
}

func GetEntryCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")
		thumbnail := c.DefaultQuery("thumbnail", "false")

		var img []byte
		var err error
		var mimetype string
		if thumbnail == "true" {
			img, mimetype, err = s.Store.GetEntryThumbnail(sid, eid)
		} else {
			img, mimetype, err = s.Store.GetEntryCoverFile(sid, eid)
		}

		if err != nil || len(img) == 0 {
			log.Debug().Err(err).Int("img length", len(img)).Str("sid", sid).Str("eid", eid).Msg("failed to get entry")
			c.AbortWithStatus(500)
			return
		}
		c.Data(200, mimetype, img)
	}
}

func PatchEntryCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		// Ensure the series exists
		_, err := s.Store.GetEntry(sid, eid)
		if err != nil {
			c.AbortWithStatusJSON(404, PatchCoverReply{Success: false})
			return
		}

		// Validate form data
		filename := c.PostForm("filename")
		file, err := c.FormFile("file")
		if err != nil || filename == "" {
			log.Debug().Err(err).Str("filename", filename).Msg("invalid form for patching cover")
			c.AbortWithStatusJSON(400, PatchCoverReply{Success: false})
			return
		}

		// Save the file
		t, err := s.Store.GetSeriesFolderTitle(sid)
		if err != nil {
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}
		imageType, err := image.InferType(filename)
		if err != nil {
			log.Debug().Err(err).Str("filename", filepath.Ext(filename)).Msg("failed getting image type for new cover")
			c.AbortWithStatusJSON(400, PatchCoverReply{Success: false})
			return
		}
		fp := fmt.Sprintf("%s/%s/.tanuki/%s.%s", s.Conf.Paths.Library, t, eid, imageType.String())
		err = fse.EnsureFileDir(fp)
		if err != nil {
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}
		err = c.SaveUploadedFile(file, fp)
		if err != nil {
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Set the new entry cover data
		cover, err := s.Store.GetEntryCover(sid, eid)
		if err != nil {
			log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not get cover")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}
		cover.Fp = fp
		cover.ImageType = imageType
		err = s.Store.SetEntryCover(sid, eid, cover)
		if err != nil {
			log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not save series cover")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Generate the new thumbnail
		if err := s.Store.GenerateEntryThumbnail(sid, eid, true); err != nil {
			log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not create thumbnail for new cover")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Reply to user
		c.JSON(200, PatchCoverReply{Success: true})
	}
}

func DeleteEntryCover(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.Param("sid")
		eid := c.Param("eid")

		// Ensure the series exists
		_, err := s.Store.GetEntry(sid, eid)
		if err != nil {
			c.AbortWithStatusJSON(404, PatchCoverReply{Success: false})
			return
		}

		err = s.Store.DeleteEntryCover(sid, eid)
		if err != nil {
			log.Debug().Err(err).Str("sid", sid).Str("eid", eid).Msg("could not delete cover and thumbnail")
			c.AbortWithStatusJSON(500, PatchCoverReply{Success: false})
			return
		}

		// Reply to user
		c.JSON(200, PatchCoverReply{Success: true})
	}
}
