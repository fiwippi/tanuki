package tags

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(api *server.RouterGroup) {
	g := api.Group("/tags")
	g.Use(cookie.Auth(api.Server))

	g.GET("/", GetTags)
	g.GET("/:tag", GetSeriesWithTag)
}
