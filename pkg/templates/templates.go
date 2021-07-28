package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fiwippi/tanuki/internal/multitemplate"
	"github.com/rs/zerolog/log"
)

var debug = false

// Global functions which can be used in all templates
var funcMap = template.FuncMap{
	// Versions files so they dont get cached (used when debugging)
	"versioning": func(filePath string) string {
		if debug {
			return fmt.Sprintf("%s?q=%s", filePath, strconv.Itoa(int(time.Now().Unix())))
		}
		return filePath
	},
}

// Renderer renders the templates from the fs
func Renderer(efs fs.FS, d bool, prefix string) multitemplate.Renderer {
	debug = d

	var r multitemplate.Renderer
	if os.Getenv("DOCKER") == "true" {
		// Always static renderer
		r = multitemplate.New()
	} else {
		// Static renderer unless gin is in debug
		// mode where it then becomes a dynamic renderer
		r = multitemplate.NewRenderer()
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

func addTemplates(layoutsDir, includesDir string, f fs.FS, r multitemplate.Renderer) (multitemplate.Renderer, error) {
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
		r.AddFromFilesFuncsFS(filepath.Base(include), funcMap, f, files...)
	}

	return r, nil
}
