package frontend

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(s *server.Server) {
	// Don't have to be authorised to login
	loginGroup := s.Group("/")
	loginGroup.Use(cookie.SkipIfAuthed(s.Session, "/"))
	loginGroup.GET("/login", login)

	// Must be authorised to access these routes
	authorised := s.Group("/")
	authorised.Use(cookie.Auth(s))
	authorised.GET("/", home)
	authorised.GET("/tags", tags)
	authorised.GET("/tags/:tag", specificTag)
	authorised.GET("/entries/:sid", entries)
	authorised.GET("/reader/:sid/:eid", reader)

	// Must be authorised and an admin to access these routes i.e. /admin
	admin := authorised.Group("/admin")
	admin.Use(cookie.Admin("/"))
	admin.GET("/", adminDashboard)
	admin.GET("/users", adminUsers)
	admin.GET("/users/edit", adminUsersEdit)
	admin.GET("/users/create", adminUsersCreate)
	admin.GET("/missing-items", adminMissingItems)
}
