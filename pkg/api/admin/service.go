package admin

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(g *server.RouterGroup) {
	a := g.Group("/admin")
	a.Use(cookie.Auth(g.Server))
	a.Use(cookie.Admin("/"))

	a.GET("/library/scan", ScanLibrary)
	a.GET("/library/generate-thumbnails", GenerateThumbnails)
	a.GET("/library/missing-items", GetMissingItems)
	a.DELETE("/library/missing-items", DeleteMissingItems)
	a.GET("/db", GetDB)
	a.GET("/users", GetUsers)
	a.PUT("/users", PutUsers)
	a.PATCH("/user/:id", PatchUser)
	a.DELETE("/user/:id", DeleteUser)
}
