package proxy

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(api *server.RouterGroup) {
	// Setup routes
	g := api.Group("/proxy")
	g.Use(cookie.Auth(api.Server))

	g.POST("/mangadex", mangadex)
}
