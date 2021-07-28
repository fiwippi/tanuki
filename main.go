package main

import (
	"embed"
	"github.com/fiwippi/tanuki/internal/encryption"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/config"
	"github.com/fiwippi/tanuki/pkg/frontend"
	"github.com/fiwippi/tanuki/pkg/logging"
	"github.com/fiwippi/tanuki/pkg/opds"
	"github.com/fiwippi/tanuki/pkg/opds/feed"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
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

// AJwPN9zUXfC675Io7edJ86zO-bF9fDcn7WlFJZPFDUA=

//go:embed files/*
var efs embed.FS

const ConfigPath = "./config/config.yml"
const SessionTTL = time.Minute * 30 // How long should a session last for
const SessionCookieName = "tanuki"  // Name of the cookie stored on the client

var opdsAuthor = &feed.Author{
	Name: "fiwippi",
	URI:  "https://github.com/fiwippi",
}

func main() {
	// Load the config
	conf := config.LoadConfig(ConfigPath)
	err := conf.Save(ConfigPath)
	if err != nil {
		log.Error().Err(err).Msg("failure to save config on startup")
	}

	// If in debug mode then set the log level to at least debug
	if conf.DebugMode && conf.Logging.Level > logging.DebugLevel {
		conf.Logging.Level = logging.DebugLevel
	}

	// Ensures the file/dir paths which tanuki uses exist
	err = conf.Paths.EnsureExist()
	if err != nil {
		log.Fatal().Err(err).Msg("paths can't be created")
	}

	// Setup the logger
	logging.SetupLogger(conf.Logging, conf.Paths.Log)

	// Create the auth session
	session := auth.NewSession(SessionTTL, SessionCookieName, *conf.SessionSecret)

	// Create the database
	db, err := bolt.Create(conf.Paths.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create db")
	}

	s := server.New(db, session, conf, opdsAuthor)

	log.Info().Msg("initialising db, this may take some time")
	scanStart := time.Now()
	err = s.ScanLibrary()
	if err != nil {
		log.Error().Err(err).Msg("failed to scan library on startup")
	} else {
		log.Info().Str("scan_time", time.Now().Sub(scanStart).String()).Msg("finished initial scan")
	}

	// Setup cron jobs
	genThumbs := func() error { return s.Store.GenerateThumbnails(true) }
	task.NewJob(conf.ScanInterval).Run(s.ScanLibrary, "scan library")
	task.NewJob(conf.ThumbGenerationInterval).Run(genThumbs, "generate thumbnails")

	// If no users exist then create default user
	if !s.Store.HasUsers() {
		pass := encryption.NewKey(32).Base64()
		err := s.Store.CreateUser(users.NewUser("default", pass, users.Admin))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create default user")
		}
		log.Info().Str("username", "default").Str("pass", pass).Msg("created default user")
	}

	// Create the router
	if !conf.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	//
	s.SetMaxMultipartMemory(int64(conf.MaxUploadedFileSizeMiB))

	// Serve static files
	files := "files"
	staticFp := files + "/static"
	templatesFp := files + "/templates"

	staticFS, err := fs.Sub(efs, staticFp)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create static filesystem")
	}
	s.Router.StaticFS("/static", http.FS(staticFS))

	// Setup the template renderer
	s.Router.HTMLRender = templates.Renderer(efs, conf.DebugMode, templatesFp)
	log.Info().Msg("templates loaded")

	// Handle 404s
	s.SetErr404(frontend.Err404(nil))

	// Routes
	api.NewService(s)
	frontend.NewService(s)
	opds.NewService(s)

	// Create the server
	srv := &http.Server{
		Addr:    conf.Host + ":" + conf.Port,
		Handler: s.Router,
	}

	log.Info().Str("host", conf.Host).Str("port", conf.Port).Str("log_level", conf.Logging.Level.String()).
		Bool("file_log", conf.Logging.LogToFile).Bool("console_log", conf.Logging.LogToConsole).Str("db_path", conf.Paths.DB).
		Str("log_path", conf.Paths.Log).Str("library_path", conf.Paths.Library).Str("mode", gin.Mode()).
		Int("max upload size", conf.MaxUploadedFileSizeMiB).Str("gin version", gin.Version).Msg("server created")

	g.Go(func() error {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server setup error")
		}
		return err
	})

	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Msg("server execution error")
	}
}
