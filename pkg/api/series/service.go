package series

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(g *server.RouterGroup) {
	series := g.Group("/series")
	series.Use(cookie.Auth(g.Server, cookie.Abort))

	series.GET("/:sid/cover", GetSeriesCover)
	series.GET("/:sid/progress", GetSeriesProgress)
	series.PATCH("/:sid/progress", PatchSeriesProgress)
	series.PATCH("/:sid/tags", PatchSeriesTags)

	entries := series.Group("/:sid/entries")
	entries.GET("/:eid/progress", GetEntryProgress)
	entries.PATCH("/:eid/progress", PatchEntryProgress)
	entries.GET("/:eid/cover", GetEntryCover)
	entries.GET("/:eid/archive", GetEntryArchive)
	entries.GET("/:eid/page/:num", GetEntryPage)
}
