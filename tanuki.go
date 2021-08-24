package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/fiwippi/tanuki/internal/pretty"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/config"
	"github.com/fiwippi/tanuki/pkg/favicon"
	"github.com/fiwippi/tanuki/pkg/frontend"
	"github.com/fiwippi/tanuki/pkg/logging"
	"github.com/fiwippi/tanuki/pkg/opds"
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
	"github.com/fiwippi/tanuki/pkg/templates"
)

var g errgroup.Group

//go:embed files/minified*
var efs embed.FS

const ConfigPath = "./config/config.yml"
const SessionTTL = time.Hour * 6   // How long should a session last for
const SessionCookieName = "tanuki" // Name of the cookie stored on the client

var opdsAuthor = &feed.Author{
	Name: "fiwippi",
	URI:  "https://github.com/fiwippi",
}

func main() {
	cfPath := flag.String("config", ConfigPath, "path to the config file, if it does not exist then it will be created")
	flag.Parse()

	// Load the config
	conf := config.LoadConfig(*cfPath)
	err := conf.Save(*cfPath)
	if err != nil {
		log.Error().Err(err).Msg("failure to save config on startup")
	}

	// Setup the logger
	logging.SetupLogger(conf.Logging, conf.Paths.Log)

	// Create the auth session
	session := auth.NewSession(SessionTTL, SessionCookieName, *conf.SessionSecret)

	// Create the database
	db, err := bolt.Startup(conf.Paths.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create db")
	}

	// Create the server
	s := server.New(db, session, conf, opdsAuthor)

	// Setup cron jobs
	conf.ScanInterval.Run(s.ScanLibrary, "scan library", true)
	go func() {
		thumbStart := time.Now()
		err = s.Store.GenerateThumbnails(false)
		log.Debug().Err(err).Str("time_taken", pretty.Duration(time.Now().Sub(thumbStart))).Msg("thumbnail generation finished")
	}()

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
	s.Router.HTMLRender = templates.CreateRenderer(s, efs, conf.DebugMode, templatesFp)
	log.Info().Msg("templates loaded")

	// Set the favicon

	// Handle 404s
	s.SetErr404(frontend.Err404(nil))

	// Routes
	api.NewService(s)
	frontend.NewService(s)
	opds.NewService(s)
	favicon.NewService(s, efs, "files/minified/static/icon/favicon.ico")

	log.Info().Str("host", conf.Host).Str("port", conf.Port).Str("log_level", conf.Logging.Level.String()).
		Bool("file_log", conf.Logging.LogToFile).Bool("console_log", conf.Logging.LogToConsole).Str("db_path", conf.Paths.DB).
		Str("log_path", conf.Paths.Log).Str("library_path", conf.Paths.Library).Str("mode", gin.Mode()).
		Int("max_upload_size", conf.MaxUploadedFileSizeMiB).Str("gin_version", gin.Version).Msg("server created")

	g.Go(func() error {
		err := s.HTTPServer().ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server setup error")
		}
		return err
	})

	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Msg("server execution error")
	}
}
