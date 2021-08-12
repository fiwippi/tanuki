package main

import (
	"embed"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/config"
	"github.com/fiwippi/tanuki/pkg/frontend"
	"github.com/fiwippi/tanuki/pkg/logging"
	"github.com/fiwippi/tanuki/pkg/opds"
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
	"github.com/fiwippi/tanuki/pkg/task"
	"github.com/fiwippi/tanuki/pkg/templates"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

//go:embed files/minified*
var efs embed.FS

const ConfigPath = "./config/config.yml"
const SessionTTL = time.Minute * 30 // How long should a session last for
const SessionCookieName = "tanuki"  // Name of the cookie stored on the client

var opdsAuthor = &feed.Author{
	Name: "fiwippi",
	URI:  "https://github.com/fiwippi",
}

// TODO better way to generate thumbnails than to always wait 30 sec each load to display theme

func main() {
	// Load the config
	conf := config.LoadConfig(ConfigPath)
	err := conf.Save(ConfigPath)
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
	genThumbs := func() error { return s.Store.GenerateThumbnails(true) }
	task.NewJob(conf.ScanInterval).Run(s.ScanLibrary, "scan library", true)
	task.NewJob(conf.ThumbGenerationInterval).Run(genThumbs, "generate thumbnails", true)

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

	// Handle 404s
	s.SetErr404(frontend.Err404(nil))

	// Routes
	api.NewService(s)
	frontend.NewService(s)
	opds.NewService(s)

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
