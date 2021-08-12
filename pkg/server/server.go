package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/config"
	"github.com/fiwippi/tanuki/pkg/logging"
	"github.com/fiwippi/tanuki/pkg/mangadex"
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
)

type Server struct {
	//
	Store    *bolt.DB
	Session  *auth.Session
	Conf     *config.Config
	Author   *feed.Author
	Router   *gin.Engine
	Mangadex *mangadex.Client

	//
	err404 gin.HandlerFunc
}

func New(store *bolt.DB, session *auth.Session, conf *config.Config, a *feed.Author) *Server {
	r := gin.New()

	// Attach basic middleware
	r.Use(logging.Middleware())
	r.Use(gin.Recovery())

	// Create the router
	if !conf.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Set the max memory size
	r.MaxMultipartMemory = int64(conf.MaxUploadedFileSizeMiB) << 20

	return &Server{
		Store:    store,
		Session:  session,
		Conf:     conf,
		Author:   a,
		Router:   r,
		Mangadex: mangadex.NewClient(),
	}
}

func (s *Server) Group(relativePath string) *RouterGroup {
	return &RouterGroup{
		RouterGroup: s.Router.Group(relativePath),
		Server:      s,
	}
}

func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:    s.Conf.Host + ":" + s.Conf.Port,
		Handler: s.Router,
	}
}
