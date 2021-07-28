package api

import (
	"github.com/fiwippi/tanuki/pkg/api/admin"
	"github.com/fiwippi/tanuki/pkg/api/auth"
	"github.com/fiwippi/tanuki/pkg/api/catalog"
	"github.com/fiwippi/tanuki/pkg/api/series"
	"github.com/fiwippi/tanuki/pkg/api/tags"
	"github.com/fiwippi/tanuki/pkg/api/user"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(s *server.Server) {
	api := s.Group("/api")

	admin.NewService(api)
	auth.NewService(api)
	catalog.NewService(api)
	series.NewService(api)
	tags.NewService(api)
	user.NewService(api)
}
