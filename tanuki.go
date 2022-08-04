package main

// TODO
// 	1. Implement interface to download manga
//  2. Implement managers for the interfaces
//  3. Implement users
//  4. Implement storage backend (solve problems like user progress + catalog metadata)
//  5. Implement OPDS
//  5. Implement frontend
//  6. Touch up frontend (e.g. more swirly loading icons in places)
//  7. Implement metadata?

// TODO go over all github issues to make sure they all covered
// TODO test on mobile after docker build made

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
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
	s := server.NewInstance(conf, store, session)

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
	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Msg("server execution error")
	}
}
