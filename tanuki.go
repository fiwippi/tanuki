package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/config"
	"github.com/fiwippi/tanuki/pkg/favicon"
	"github.com/fiwippi/tanuki/pkg/frontend"
	"github.com/fiwippi/tanuki/pkg/opds"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/storage"
	"github.com/fiwippi/tanuki/pkg/templates"
	"github.com/fiwippi/tanuki/pkg/transfer"
)

//go:embed files/minified*
var efs embed.FS

func main() {
	recreate := flag.Bool("recreate", false, "recreate the db on startup")
	cfPath := flag.String("config", "./config/config.yml", "path to the config file, if it does not exist then it will be created")
	flag.Parse()

	// Load the config
	conf := config.LoadConfig(*cfPath)
	if err := conf.Save(*cfPath); err != nil {
		log.Error().Err(err).Msg("failure to save config on startup")
	}

	// Setup the logger
	log.Setup(conf.Logging.Level,
		conf.Logging.LogToConsole,
		conf.Logging.LogToFile,
		conf.Paths.Log)

	// Create the server
	session := auth.NewSession(time.Hour*24*3, "tanuki", *conf.SessionSecret)
	store := storage.MustCreateNewStore(conf.Paths.DB, conf.Paths.Library, *recreate)
	manager := transfer.NewManager(conf.Paths.Library, 2, store,
		store.PopulateCatalog, conf.SubscriptionsInterval.Duration)
	s := server.NewInstance(conf, store, session, manager)

	// Serve static files
	files := "files/minified"
	staticFp := files + "/static"
	templatesFp := files + "/templates"

	staticFS, err := fs.Sub(efs, staticFp)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create static filesystem")
	}
	s.Router.StaticFS("/static", http.FS(staticFS))

	// Setup the template renderer
	templates.CreateRenderer(s, efs, conf.DebugMode, templatesFp)
	log.Info().Msg("templates loaded")

	// Register routes
	favicon.NewService(s, efs, "files/minified/static/icon/favicon.ico")
	api.NewService(s)
	frontend.NewService(s)
	opds.NewService(s)

	log.Info().Str("host", conf.Host).Str("port", conf.Port).Str("log_level", conf.Logging.Level.String()).
		Bool("file_log", conf.Logging.LogToFile).Bool("console_log", conf.Logging.LogToConsole).Str("db_path", conf.Paths.DB).
		Str("log_path", conf.Paths.Log).Str("library_path", conf.Paths.Library).Str("mode", gin.Mode()).
		Int("max_upload_size", conf.MaxUploadedFileSizeMiB).Str("gin_version", gin.Version).Msg("server created")

	var g errgroup.Group
	g.Go(func() error {
		err := s.Start()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server setup error")
		}
		return err
	})
	g.Go(func() error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
		defer cancel()
		<-ctx.Done()
		return s.Shutdown()
	})
	if err := g.Wait(); err != nil {
		if err == http.ErrServerClosed {
			log.Info().Msg("server closed successfully")
		} else {
			log.Fatal().Err(err).Msg("server execution error")
		}
	}
}
