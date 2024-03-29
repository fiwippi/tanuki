package auth

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(g *server.RouterGroup) {
	authGroup := g.Group("/auth")

	// Don't have to be authorised to login
	authGroup.POST("/login", login)

	// Must be authorised to logout
	authorised := authGroup.Use(cookie.Auth(g.Server, cookie.Abort))
	authorised.GET("/logout", logout)
}
