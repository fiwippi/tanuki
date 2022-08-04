package user

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

// These are api routes each user can use to get and
// modify properties about themselves, users are
// expected to provide the cookie identifying themselves
// in order to access/edit their own data unless they're
// an admin

func NewService(api *server.RouterGroup) {
	g := api.Group("/user")
	g.Use(cookie.Auth(api.Server, cookie.Abort))

	g.GET("/type", GetType)
	g.GET("/name", GetName)
}
