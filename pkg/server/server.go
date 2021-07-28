package server

import (
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/config"
	"github.com/fiwippi/tanuki/pkg/logging"
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
	"github.com/gin-gonic/gin"
)

type Server struct {
	//
	Store   *bolt.DB
	Session *auth.Session
	Conf    *config.Config
	Author  *feed.Author
	Router  *gin.Engine

	//
	err404 gin.HandlerFunc
}

func New(store *bolt.DB, session *auth.Session, conf *config.Config, a *feed.Author) *Server {
	r := gin.New()

	// Attach basic middleware
	r.Use(logging.Middleware())
	r.Use(gin.Recovery())

	return &Server{
		Store:   store,
		Session: session,
		Conf:    conf,
		Author:  a,
		Router:  r,
	}
}

func (s *Server) SetMaxMultipartMemory(sizeMiB int64) {
	s.Router.MaxMultipartMemory = sizeMiB << 20
}

func (s *Server) Group(relativePath string) *RouterGroup {
	return &RouterGroup{
		RouterGroup: s.Router.Group(relativePath),
		Server:      s,
	}
}
