package templates

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/multitemplate"
	"github.com/fiwippi/tanuki/pkg/server"
)

// CreateRenderer creates a Renderer which  renders the templates from the fs
func CreateRenderer(s *server.Server, efs fs.FS, debug bool, prefix string) multitemplate.Renderer {
	var temp multitemplate.Renderer
	if os.Getenv("DOCKER") == "true" {
		// Always static renderer
		temp = multitemplate.New()
	} else {
		// Static renderer unless gin is in debug
		// mode where it then becomes a dynamic renderer
		temp = multitemplate.NewRenderer()
	}

	r := &Renderer{
		Renderer: temp,
		server:   s,
		debug:    debug,
	}

	// Generating our main templates
	newR, err := addTemplates(prefix+"/layouts/base.tmpl", prefix+"/includes/*.tmpl", efs, r)
	if err == nil {
		r = newR
	}

	// Generating our templates which do not need a header
	newR, err = addTemplates(prefix+"/layouts/blank_base.tmpl", prefix+"/blank_includes/*.tmpl", efs, r)
	if err == nil {
		r = newR
	}

	return r
}

func addTemplates(layoutsDir, includesDir string, f fs.FS, r *Renderer) (*Renderer, error) {
	layouts, err := fs.Glob(f, layoutsDir)
	if err != nil {
		log.Error().Err(err).Str("dir", layoutsDir).Msg("could not get layouts dir")
		return nil, err
	}
	includes, err := fs.Glob(f, includesDir)
	if err != nil {
		log.Error().Err(err).Str("dir", includesDir).Msg("could not get includes dir")
		return nil, err
	}

	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		r.AddFromFilesFuncsFS(filepath.Base(include), r.FuncMap(), f, files...)
	}

	return r, nil
}
