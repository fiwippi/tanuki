package api

import (
	"github.com/fiwippi/tanuki/pkg/api/admin"
	"github.com/fiwippi/tanuki/pkg/api/auth"
	"github.com/fiwippi/tanuki/pkg/api/manager"
	"github.com/fiwippi/tanuki/pkg/api/mangadex"
	"github.com/fiwippi/tanuki/pkg/api/series"
	"github.com/fiwippi/tanuki/pkg/api/user"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(s *server.Instance) {
	api := s.Group("/api")

	admin.NewService(api)
	auth.NewService(api)
	series.NewService(api)
	user.NewService(api)
	mangadex.NewService(api)
	manager.NewService(api)
}
