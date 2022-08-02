package opds

import (
	"github.com/fiwippi/tanuki/pkg/auth/basic"
	"github.com/fiwippi/tanuki/pkg/server"
)

const opdsRoot = "/opds/v1.2"
const authorName = "fiwippi"
const authorURI = "https://github.com/fiwippi"

func NewService(s *server.Instance) {
	opds := s.Group("/opds")
	opds.Use(basic.Auth("Tanuki OPDS", s.Store))

	v1p2 := opds.Group("/v1.2")
	v1p2.GET("/catalog", GetCatalog)
	v1p2.GET("/series/:sid", GetEntries)
	v1p2.GET("/series/:sid/entries/:eid/archive", GetArchive)
	v1p2.GET("/series/:sid/entries/:eid/cover", GetCover)
	v1p2.GET("/series/:sid/entries/:eid/page/:num", GetPage)
}
