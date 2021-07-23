package tanuki

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/core"
	"github.com/fiwippi/tanuki/pkg/opds"
)

var opdsAuthor = &opds.Author{
	Name: "fiwippi",
	URI:  "https://github.com/fiwippi",
}

// GET /opds/v1.2/catalog
func opdsCatalog(c *gin.Context) {
	catalog := opds.NewCatalog()
	catalog.SetAuthor(opdsAuthor)

	updated, err := db.GetCatalogModTime()
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	catalog.SetUpdated(updated)

	list := db.GetCatalog()
	for _, s := range list {
		seriesTime, err := db.GetSeriesModTime(s.Hash)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		catalog.AddEntry(&opds.SeriesEntry{
			Title:   s.Title,
			Updated: opds.Time{seriesTime},
			ID:      s.Hash,
		})
	}

	c.XML(200, catalog)
}

// GET /opds/v1.2/series/:sid
func opdsViewEntries(c *gin.Context) {
	id := c.Param("sid")
	if _, err := db.GetSeries(id); err != nil {
		c.AbortWithStatus(404)
		return
	}

	data, err := db.GetSeries(id)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	series := opds.NewSeries(data.Hash, data.Title)
	series.SetAuthor(opdsAuthor)
	updated, err := db.GetSeriesModTime(id)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	series.SetUpdated(updated)

	list, err := db.GetEntries(id)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	for _, e := range list {
		entryTime, err := db.GetEntryModTime(id, e.Hash)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		cover, err := db.GetEntryCover(id, e.Hash)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		a, err := db.GetEntryArchive(id, e.Hash)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		series.AddEntry(&opds.ArchiveEntry{
			Title:     e.Title,
			Updated:   opds.Time{entryTime},
			ID:        e.Hash,
			CoverType: cover.ImageType,
			ThumbType: core.ImageJPEG,
			Archive:   a,
		})
	}

	c.XML(200, series)
}

// GET /opds/v1.2/series/:sid/entries/:eid/archive
func opdsArchive(c *gin.Context) {
	apiGetEntryArchive(c)
}

// GET /opds/v1.2/series/:sid/entries/:eid/cover?thumbnail={true,false}
func opdsCover(c *gin.Context) {
	apiGetEntryCover(c)
}
