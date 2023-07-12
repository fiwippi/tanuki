package mangadex

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(g *server.RouterGroup) {
	mdex := g.Group("/mangadex")
	mdex.Use(cookie.Auth(g.Server, cookie.Abort))

	mdex.GET("/search", searchMangadex)
	mdex.GET("/view/:uuid", viewManga)
	mdex.GET("/cover/:endpoint", getCover)
}
