package main

import (
	"embed"
	"net/http"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/fiwippi/tanuki/pkg/tanuki"
)

var g errgroup.Group

// 4cGRXSkCMcAMRLkH1fRx-ug_7JjpLz1_o_ihBsUBvPs=

//go:embed files/*
var f embed.FS

func main() {
	g.Go(func() error {
		err := tanuki.Server(f).ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server setup error")
		}
		return err
	})

	if err := g.Wait(); err != nil {
		log.Fatal().Err(err).Msg("server execution error")
	}
}
