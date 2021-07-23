package tanuki

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
	database "github.com/fiwippi/tanuki/pkg/db"
)

var conf *config             // Current config which tanuki is using
var db *database.DB          // The user/manga database
var static, templates string // Filepaths to the static dir and templates dir

func init() {
	files := "files"
	static = files + "/static"
	templates = files + "/templates"
}

// Server sets up the server in with the following steps:
// 1. load up the config.yml file
// 2. ensure paths directories exist, tanuki uses them to store/serve data
// 3. setup the log level and outputs
// 4. create/load the database, create default user if none exist
// 5. create the auth session to store cookies to validate users
// 6. create the router and attach it to a *http.Server
// efs is the embedded file system found in files/ in order to serve
// the templates and other static files
func Server(efs fs.FS) *http.Server {
	// Load the config file
	conf = loadConfig()
	err := saveConfig(conf)
	if err != nil {
		log.Error().Err(err).Msg("failure to save config on startup")
	}

	// Debug mode must at least log in debug
	if conf.DebugMode && conf.Logging.Level > DebugLevel {
		conf.Logging.Level = DebugLevel
	}

	// Ensures the file/dir paths which tanuki uses exist
	err = conf.Paths.EnsureExist()
	if err != nil {
		log.Fatal().Err(err).Msg("paths can't be created")
	}

	// Setup the logger
	setupLogger(conf.Logging, conf.Paths.Log)

	// Create the database and if no users exist then create default user
	db, err = database.CreateDB(conf.Paths.DB)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create db")
	}
	log.Info().Msg("initialising db, this may take some time")
	scanStart := time.Now()
	err = ScanLibrary()
	if err != nil {
		log.Error().Err(err).Msg("failed to scan library on startup")
	} else {
		log.Info().Str("scan_time", time.Now().Sub(scanStart).String()).Msg("finished scan")
	}
	thumbnailStart := time.Now()
	err = db.GenerateThumbnails(true)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate thumbnails startup")
	} else {
		log.Info().Str("task_time", time.Now().Sub(thumbnailStart).String()).Msg("finished thumbnail generation")
	}

	// Setup cron jobs
	conf.ScanIntervalMinutes.RunTask(ScanLibrary, "scan library")
	NewInterval(10).RunTask(func() error { return db.GenerateThumbnails(true) }, "generate thumbnails")

	if !db.HasUsers() {
		pass := auth.NewSecureKey(32).Base64()
		err := db.CreateUser(core.NewUser("default", pass, core.AdminUser))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create default user")
		}
		log.Info().Str("username", "default").Str("pass", pass).Msg("created default user")
	}

	// Create the session
	session = auth.NewSession(authTime, authCookieName, *conf.SessionSecret)

	// Create the router
	if !conf.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := createRouter(efs)

	// Attach the router to a http server
	srv := &http.Server{
		Addr:    conf.Host + ":" + conf.Port,
		Handler: r,
	}

	log.Info().Str("host", conf.Host).Str("port", conf.Port).Str("log_level", conf.Logging.Level.String()).
		Bool("file_log", conf.Logging.LogToFile).Bool("console_log", conf.Logging.LogToConsole).Str("db_path", conf.Paths.DB).
		Str("log_path", conf.Paths.Log).Str("library_path", conf.Paths.Library).Str("mode", gin.Mode()).
		Int("max upload size", conf.MaxUploadedFileSizeMiB).Str("gin version", gin.Version).Msg("server created")

	return srv
}

// Resident in england more than 3 years prior
// Letter from place of education, dates of attendance, flight ticket, signed letter on heavy paper, work permits?, flight details
