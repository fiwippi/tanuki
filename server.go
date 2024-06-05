package tanuki

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server Config

type ServerConfig struct {
	Host         string   `json:"host"`
	HttpPort     uint16   `json:"http_port"`
	RpcPort      uint     `json:"rpc_port"`
	DataPath     string   `json:"data_path"`
	LibraryPath  string   `json:"library_path"`
	ScanInterval duration `json:"scan_interval"`
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:         "0.0.0.0",
		HttpPort:     8001,
		RpcPort:      9001,
		DataPath:     "./data",
		LibraryPath:  "./library",
		ScanInterval: duration{1 * time.Hour},
	}
}

// Server

type Server struct {
	// Tanuki data
	config ServerConfig
	store  *Store

	// Public/private endpoints
	public   *http.Server
	privateL net.Listener
	privateH *rpc.Server

	// Long-running tasks accounting
	stopScan      chan struct{}
	ackStopScan   chan struct{}
	stopVacuum    chan struct{}
	ackStopVacuum chan struct{}

	// Startup/Shutdown accounting
	started    atomic.Bool
	inShutdown atomic.Bool
	shutdown   atomic.Bool
}

func NewServer(config ServerConfig) (*Server, error) {
	if err := os.MkdirAll(config.DataPath, 0666); err != nil {
		return nil, err
	}
	store, err := NewStore(filepath.Join(config.DataPath, "store"))
	if err != nil {
		return nil, err
	}

	s := &Server{
		config: config,
		store:  store,
		public: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", config.Host, config.HttpPort),
			Handler: router(store),
		},
		privateH:      rpc.NewServer(),
		stopScan:      make(chan struct{}),
		ackStopScan:   make(chan struct{}),
		stopVacuum:    make(chan struct{}),
		ackStopVacuum: make(chan struct{}),
	}
	if err := s.privateH.Register(s); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Start() error {
	privateAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.RpcPort)
	slog.Info("Starting server",
		slog.Any("public", s.public.Addr),
		slog.Any("private", privateAddr))

	// Start the public OPDS server
	l, err := net.Listen("tcp", s.public.Addr)
	if err != nil {
		return err
	}
	defer s.started.Store(true)
	go func() {
		if err := s.public.Serve(l); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Server execution failed", slog.Any("err", err))
				os.Exit(1)
			}
		}
	}()

	// Start the private RPC listener
	s.privateL, err = net.Listen("tcp", privateAddr)
	if err != nil {
		slog.Error("RPC listen failed", slog.Any("err", err))
		return err
	}
	go func() {
		for {
			conn, err := s.privateL.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					slog.Error("Failed to accept incoming RPC", slog.Any("err", err))
				}
				break
			}
			slog.Info("Accepted RPC", slog.Any("address", conn.RemoteAddr()))

			go func() {
				defer conn.Close()
				s.privateH.ServeConn(conn)
			}()
		}
	}()

	// Start long-running tasks
	go s.scan()
	go s.vacuum()

	return nil
}

func (s *Server) Stop() {
	// Don't attempt to shut down if the server
	// hasn't started the listener and goroutines
	if !s.started.Load() {
		return
	}

	// If we've already shutdown then exit
	if s.shutdown.Load() {
		slog.Warn("Server already shut down")
		return
	}

	// If we are already shutting down in another
	// goroutine then wait until we have shut down
	if !s.inShutdown.CompareAndSwap(false, true) {
		slog.Warn("Server already shutting down")
		for !s.shutdown.Load() {
			time.Sleep(50 * time.Millisecond)
		}
		return
	}

	defer func() {
		s.shutdown.Store(true)
		slog.Info("Server stopped")
	}()

	// Stop long-running tasks
	s.stopScan <- struct{}{}
	<-s.ackStopScan
	close(s.stopScan)
	close(s.ackStopScan)
	s.stopVacuum <- struct{}{}
	<-s.ackStopVacuum
	close(s.stopVacuum)
	close(s.ackStopVacuum)

	// Stop the server
	//
	// We panic because an errored shutdown
	// puts the server in an invalid state
	// which we can't recover from
	if err := s.public.Close(); err != nil {
		panic(err)
	}
	if err := s.privateL.Close(); err != nil {
		panic(err)
	}
}

func router(s *Store) *chi.Mux {
	r := chi.NewRouter()
	r.Use(httpLogger())

	// Public routes
	r.Route(opdsRoot, func(r chi.Router) {
		r.Use(basicAuth("Tanuki OPDS", s))

		r.Get("/search", handleSearch())
		r.Get("/catalog", handleCatalog(s))
		r.Get("/series/{sid}", handleEntries(s))
		r.Get("/series/{sid}/entries/{eid}/archive", handleArchive(s))
		r.Get("/series/{sid}/entries/{eid}/cover", handleCover(s))
		r.Get("/series/{sid}/entries/{eid}/page/{num}", handlePage(s))
	})

	return r
}

// Tasks

func (s *Server) scan() {
	task := func() {
		slog.Info("Scanning library", slog.String("path", s.config.LibraryPath))

		start := time.Now()

		lib, err := ParseLibrary(s.config.LibraryPath)
		if err != nil {
			slog.Error("Failed to (fully) scan library", slog.Any("err", err))
		}

		if len(lib) >= 1 {
			if err := s.store.PopulateCatalog(lib); err != nil {
				slog.Error("Failed to populate catalog", slog.Any("err", err))
			} else {
				timeTaken := time.Since(start).Round(time.Millisecond)
				slog.Info("Scanned library", slog.Duration("duration", timeTaken))
			}
		}
	}

	t := time.NewTicker(s.config.ScanInterval.Duration)

	task() // We want to scan on startup
	for {
		select {
		case <-t.C:
			task()
		case <-s.stopScan:
			t.Stop()
			slog.Info("Done scanning")
			s.ackStopScan <- struct{}{}
			return
		}
	}
}

func (s *Server) vacuum() {
	t := time.NewTicker(24 * time.Hour)

	for {
		select {
		case <-t.C:
			slog.Info("Vacuuming store")
			if err := s.store.Vacuum(); err != nil {
				slog.Error("Failed to vacuum store", slog.Any("err", err))
			} else {
				slog.Info("Vacuumed store")
			}
		case <-s.stopVacuum:
			t.Stop()
			slog.Info("Done vacuuming")
			s.ackStopVacuum <- struct{}{}
			return
		}
	}
}

// OPDS Handlers

const opdsMime = "application/atom+xml"

func handleSearch() http.HandlerFunc {
	// Pre-encode the search-related XML
	// since it remains static
	encodedSearch, err := xml.MarshalIndent(newOpdsSearch(), "", xmlIndent)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write(encodedSearch)
	}
}

func handleCatalog(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalog, err := s.GetCatalog()
		if err != nil {
			slog.Error("Failed to retrieve catalog", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var modTime time.Time
		for _, series := range catalog {
			if series.ModTime.After(modTime) {
				modTime = series.ModTime
			}
		}

		c := newOpdsFeed("ctl", "Catalog", modTime, opdsAuthor{
			Name: "fiwippi",
			URI:  "https://github.com/fiwippi",
		})
		c.addLink("/catalog", relSelf, typeNavigation)
		c.addLink("/search", relSearch, typeSearch)

		filter := r.URL.Query().Get("search")
		for _, series := range catalog {
			if len(filter) > 0 && !fuzzy(series.Title, filter) {
				continue
			}
			c.addSeries(&series)
		}

		w.Header().Set("Content-Type", opdsMime)
		w.WriteHeader(http.StatusOK)
		if err := newXmlEncoder(w).Encode(c); err != nil {
			slog.Error("Failed to encode catalog", slog.Any("err", err))
			return
		}
	}
}

func handleEntries(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		if sid == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		series, err := s.GetSeries(sid)
		if err != nil {
			slog.Error("Failed to retrieve series", slog.Any("err", err), slog.String("sid", sid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		entries, err := s.GetEntries(sid)
		if err != nil {
			slog.Error("Failed to retrieve entries", slog.Any("err", err), slog.String("sid", sid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c := newOpdsFeed(series.SID, series.Title, series.ModTime, opdsAuthor{Name: series.Author})
		c.addLink("/series/"+series.SID, relSelf, typeAcquisition)
		for _, e := range entries {
			c.addEntry(&e)
		}

		w.Header().Set("Content-Type", opdsMime)
		w.WriteHeader(http.StatusOK)
		if err := newXmlEncoder(w).Encode(c); err != nil {
			slog.Error("Failed to encode entries", slog.Any("err", err), slog.String("sid", sid))
			return
		}
	}
}

func handleArchive(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		eid := r.PathValue("eid")
		if sid == "" || eid == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		entry, err := s.GetEntry(sid, eid)
		if err != nil {
			slog.Error("Failed to retrieve entry", slog.Any("err", err),
				slog.String("sid", sid), slog.String("eid", eid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sendFileAsAttachment(w, r, entry.Archive)
	}
}

func handleCover(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		eid := r.PathValue("eid")
		if sid == "" || eid == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		thumbnail := r.URL.Query().Get("thumbnail") == "true"

		var err error
		var mime string
		var cover *bytes.Buffer
		if thumbnail {
			cover, mime, err = s.GetThumbnail(sid, eid)
		} else {
			cover, mime, err = s.GetPage(sid, eid, 0)
		}
		if err != nil {
			slog.Error("Failed to retrieve cover", slog.Any("err", err),
				slog.String("sid", sid), slog.String("eid", eid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sendFile(w, cover, mime)
	}
}

func handlePage(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		eid := r.PathValue("eid")
		if sid == "" || eid == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		num, err := strconv.Atoi(r.PathValue("num"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		page, mime, err := s.GetPage(sid, eid, num)
		if err != nil {
			slog.Error("Failed to retrieve page", slog.Any("err", err),
				slog.String("sid", sid), slog.String("eid", eid), slog.Int("num", num))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sendFile(w, page, mime)
	}
}

// RPCs

func (s *Server) Scan(_ struct{}, _ *struct{}) error {
	slog.Info("Manually scanning library")
	lib, err := ParseLibrary(s.config.LibraryPath)
	if err != nil && len(lib) == 0 {
		return err
	}

	if len(lib) >= 1 {
		if err != nil {
			slog.Error("Failed to (fully) manually scan library", slog.Any("err", err))
		}
		if err := s.store.PopulateCatalog(lib); err != nil {
			slog.Error("Failed to populate (manually scanned) catalog", slog.Any("err", err))
			return err
		}
		slog.Info("Manual scan complete")
	}

	return nil
}

func (s *Server) Dump(_ struct{}, output *string) error {
	slog.Info("Dumping store")
	out, err := s.store.Dump()
	if err != nil {
		slog.Error("Failed to dump store", slog.Any("err", err))
		return err
	}
	slog.Info("Dumped store")
	*output = out
	return nil
}

func (s *Server) AddUser(u User, _ *struct{}) error {
	log := slog.With(slog.String("name", u.Name))

	log.Info("Adding user")
	if err := s.store.AddUser(u.Name, u.Pass); err != nil {
		log.Error("Failed to add user", slog.Any("err", err))
		return err
	}
	log.Info("Added user")
	return nil
}

func (s *Server) DeleteUser(name string, _ *struct{}) error {
	log := slog.With(slog.String("name", name))

	log.Info("Deleting user")
	if err := s.store.DeleteUser(name); err != nil {
		log.Error("Failed to delete user", slog.Any("err", err))
		return err
	}
	slog.Info("Deleted user")
	return nil
}

type ChangeUsernameRequest struct {
	OldName, NewName string
}

func (s *Server) ChangeUsername(req ChangeUsernameRequest, _ *struct{}) error {
	log := slog.With(slog.String("old", req.OldName), slog.String("new", req.NewName))

	log.Info("Changing username")
	if err := s.store.ChangeUsername(req.OldName, req.NewName); err != nil {
		log.Error("Failed to change username", slog.Any("err", err))
		return err
	}
	log.Info("Changed username")
	return nil
}

type ChangePasswordRequest struct {
	Name, Password string
}

func (s *Server) ChangePassword(req ChangePasswordRequest, _ *struct{}) error {
	log := slog.With(slog.String("name", req.Name))

	log.Info("Changing password")
	if err := s.store.ChangePassword(req.Name, req.Password); err != nil {
		log.Error("Failed to change password", slog.Any("err", err))
		return err
	}
	log.Info("Changed password")
	return nil
}

// Helpers

func sendFile(w http.ResponseWriter, f *bytes.Buffer, mime string) {
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.FormatInt(int64(f.Len()), 10))
	io.Copy(w, f)
}

func sendFileAsAttachment(w http.ResponseWriter, r *http.Request, path string) {
	filename := url.QueryEscape(filepath.Base(path))
	w.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+filename)
	http.ServeFile(w, r, path)
}

// Logging

func httpLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			t1 := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}

				attrs := []any{
					slog.Attr{Key: "url", Value: slog.StringValue(fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI))},
					slog.Attr{Key: "method", Value: slog.StringValue(r.Method)},
					slog.Attr{Key: "ip", Value: slog.StringValue(r.RemoteAddr)},
					slog.Attr{Key: "elapsed", Value: slog.DurationValue(time.Since(t1).Round(time.Millisecond))},
				}

				slog.With(attrs...).Log(context.Background(), statusLevel(ww.Status()), fmt.Sprintf("%d", ww.Status()))
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func statusLevel(status int) slog.Level {
	switch {
	case status <= 0:
		return slog.LevelWarn
	case status < 400: // For codes in 100s, 200s, 300s
		return slog.LevelInfo
	case status >= 400 && status < 500:
		// Switching to info level to be less noisy
		return slog.LevelInfo
	case status >= 500:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Basic Authentication

var (
	errAuthHeaderEmpty   = errors.New("auth header is empty")
	errInvalidAuthFormat = errors.New("auth header formatted in incorrect way")
)

func basicAuth(realm string, store *Store) func(next http.Handler) http.Handler {
	if realm == "" {
		realm = "Authorisation Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, err := parseBasicAuthCred(r)
			if err != nil {
				slog.Debug("Parsing basic auth credentials failed", slog.Any("err", err))
				w.Header().Set("WWW-Authenticate", realm)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			valid := store.AuthLogin(user, pass)
			if !valid {
				slog.Debug("Invalid login credentials", slog.Any("err", err))
				w.Header().Set("WWW-Authenticate", realm)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parseBasicAuthCred(r *http.Request) (string, string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", "", errAuthHeaderEmpty
	}
	if !strings.HasPrefix(h, "Basic ") {
		return "", "", errInvalidAuthFormat
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(h, "Basic "))
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		return "", "", errInvalidAuthFormat
	}

	return parts[0], parts[1], nil
}

// Fuzzy Searching

func fuzzy(text, searchTerm string) bool {
	searchTerm = strings.ToUpper(searchTerm)
	text = strings.ToUpper(text)

	j := -1
	for i := 0; i < len(searchTerm); i++ {
		l := searchTerm[i]
		if l == ' ' { // Ignore spaces
			continue
		}

		j = indexOf(text, l, j+1) // Search for character and update position
		if j == -1 {
			return false
		}
	}
	return true
}

func indexOf(search string, letter uint8, start int) int {
	for i := start; i < len(search); i++ {
		if search[i] == letter {
			return i
		}
	}
	return -1
}

// XML Encoding

var xmlIndent = "" // We set indenting for tests

func newXmlEncoder(w io.Writer) *xml.Encoder {
	e := xml.NewEncoder(w)
	e.Indent("", xmlIndent)
	return e
}

// Custom duration encoding

type duration struct {
	time.Duration
}

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *duration) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	d.Duration, err = time.ParseDuration(s)
	return err
}
