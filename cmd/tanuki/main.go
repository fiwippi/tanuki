package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/pelletier/go-toml/v2"

	"github.com/fiwippi/tanuki/v2"
)

var logLevel = &slog.LevelVar{}

func init() {
	// Until we parse the log level we only emit error logs
	logLevel.Set(slog.LevelError)
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.DateTime,
		}),
	))
}

func main() {
	configPath := flag.String("config", "", "Path to config.json file. Leave blank to use the default config")
	printVersion := flag.Bool("version", false, "Output version information and exit")
	flag.Parse()

	if err := run(*configPath, *printVersion); err != nil {
		slog.Error("Failed to run tanuki", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(configPath string, printVersion bool) error {
	if printVersion {
		fmt.Printf("tanuki %s\n", tanuki.Version)
		return nil
	}

	config := tanuki.DefaultServerConfig()
	if configPath != "" {
		f, err := os.Open(configPath)
		if err != nil {
			return fmt.Errorf("open config: %w", err)
		}
		if err := toml.NewDecoder(f).Decode(&config); err != nil {
			return fmt.Errorf("decode config: %w", err)
		}
	}

	// No non-ERROR logging should happen until we set the log level!
	switch strings.ToUpper(config.LogLevel) {
	case "DEBUG":
		logLevel.Set(slog.LevelDebug)
	case "INFO":
		logLevel.Set(slog.LevelInfo)
	case "WARN":
		logLevel.Set(slog.LevelWarn)
	case "ERROR":
		logLevel.Set(slog.LevelError)
	default:
		return fmt.Errorf("invalid log level: %s", config.LogLevel)
	}

	slog.Info("Using config", slog.Any("config", config))

	s, err := tanuki.NewServer(config)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	if err := s.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-done

	s.Stop()

	return nil
}
