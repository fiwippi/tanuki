package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	zlog "github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/config"
	"github.com/fiwippi/tanuki/pkg/storage"
	"github.com/fiwippi/tanuki/pkg/transfer"
)

type Instance struct {
	Store   *storage.Store
	Session *auth.Session
	Config  *config.Config
	Router  *gin.Engine
	Manager *transfer.Manager

	srv *http.Server
}

func NewInstance(c *config.Config, store *storage.Store, session *auth.Session, m *transfer.Manager) *Instance {
	r := gin.New()
	r.Use(log.Middleware())
	r.Use(gin.Recovery())
	r.MaxMultipartMemory = int64(c.MaxUploadedFileSizeMiB) << 20

	if !c.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	return &Instance{
		Store:   store,
		Session: session,
		Config:  c,
		Router:  r,
		Manager: m,
	}
}

func (i *Instance) Group(relativePath string) *RouterGroup {
	return &RouterGroup{
		RouterGroup: i.Router.Group(relativePath),
		Server:      i,
	}
}

func (i *Instance) SetHTMLRenderer(r render.HTMLRender) {
	i.Router.HTMLRender = r
}

func (i *Instance) Shutdown() error {
	if i.srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	err := i.Store.Close()
	if err != nil {
		return err
	}
	return i.srv.Shutdown(ctx)
}

func (i *Instance) Start() error {
	// Scan on intervals
	go func() {
		err := i.Store.PopulateCatalog()
		if err != nil {
			zlog.Error().Err(err).Msg("failed to scan library on startup")
		} else {
			zlog.Info().Msg("scanned library on startup")
		}

		ticker := time.NewTicker(time.Duration(i.Config.ScanInterval) * time.Minute)
		for range ticker.C {
			err := i.Store.PopulateCatalog()
			if err != nil {
				zlog.Error().Err(err).Msg("failed to scan library on interval interval")
			} else {
				zlog.Info().Msg("scanned library on interval")
			}
		}
	}()

	// Generate thumbnails once at first load for all ones which don't yet exist
	go func() {
		zlog.Info().Msg("beginning thumbnail generation at startup")
		thumbStart := time.Now()
		err := i.Store.GenerateThumbnails(false)
		zlog.Debug().Err(err).Str("time_taken", time.Since(thumbStart).String()).
			Msg("thumbnail generation finished")
	}()

	// Run the server
	i.srv = &http.Server{
		Addr:    i.Config.Host + ":" + i.Config.Port,
		Handler: i.Router,
	}
	return i.srv.ListenAndServe()
}
