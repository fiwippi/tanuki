package catalog

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(api *server.RouterGroup) {
	g := api.Group("/catalog")
	g.Use(cookie.Auth(api.Server))

	g.GET("/", GetCatalog)
	g.GET("/progress", GetProgress)
}
