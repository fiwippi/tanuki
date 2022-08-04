package api

import (
	"github.com/fiwippi/tanuki/pkg/api/admin"
	"github.com/fiwippi/tanuki/pkg/api/auth"
	"github.com/fiwippi/tanuki/pkg/api/user"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(s *server.Instance) {
	api := s.Group("/api")

	admin.NewService(api)
	auth.NewService(api)
	//series.NewService(api)
	user.NewService(api)
	//download.NewService(api)
	//proxy.NewService(api)
}
