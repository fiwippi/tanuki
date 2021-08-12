package opds

import (
	"github.com/fiwippi/tanuki/pkg/auth/basic"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(s *server.Server) {
	opds := s.Group("/opds")
	opds.Use(basic.Auth("Tanuki OPDS", s.Store))

	v1p2 := opds.Group("/v1.2")
	v1p2.GET("/search", GetSearch)
	v1p2.GET("/catalog", GetCatalog)
	v1p2.GET("/series/:sid", GetEntries)
	v1p2.GET("/series/:sid/entries/:eid/archive", GetArchive)
	v1p2.GET("/series/:sid/entries/:eid/cover", GetCover)
	v1p2.GET("/series/:sid/entries/:eid/page/:num", GetPage)
}
