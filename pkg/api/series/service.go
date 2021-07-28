package series

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(g *server.RouterGroup) {
	series := g.Group("/series")
	series.Use(cookie.Auth(g.Server))

	series.GET("/", GetCatalog) // TODO move this route to the catalog endpoint
	series.GET("/:sid", GetSeries)
	series.PATCH("/:sid", PatchSeries)
	series.GET("/:sid/cover", GetSeriesCover)
	series.PATCH("/:sid/cover", PatchSeriesCover)
	series.DELETE("/:sid/cover", DeleteSeriesCover)
	series.GET("/:sid/progress", GetSeriesProgress)
	series.PATCH("/:sid/progress", PatchSeriesProgress)
	series.GET("/:sid/tags", GetSeriesTags)
	series.PATCH("/:sid/tags", PatchSeriesTags)

	entries := series.Group("/:sid/entries")
	entries.GET("/", GetEntries)
	entries.GET("/:eid/progress", GetEntryProgress)
	entries.PATCH("/:eid/progress", PatchEntryProgress)
	entries.GET("/:eid", GetEntry)
	entries.PATCH("/:eid", PatchEntry)
	entries.GET("/:eid/cover", GetEntryCover)
	entries.PATCH("/:eid/cover", PatchEntryCover)
	entries.DELETE("/:eid/cover", DeleteEntryCover)
	entries.GET("/:eid/archive", GetEntryArchive)
	entries.GET("/:eid/page/:num", GetEntryPage)
}
