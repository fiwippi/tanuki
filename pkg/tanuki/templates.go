package tanuki

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

// Global functions which can be used in all templates
var templateFuncMap = template.FuncMap{
	// Versions files so they dont get cached
	"versioning": func(filePath string) string {
		if conf.DebugMode {
			return fmt.Sprintf("%s?q=%s", filePath, strconv.Itoa(int(time.Now().Unix())))
		}
		return filePath
	},
}

// Load templates on program initialisation
func templateRenderer(efs fs.FS) multitemplate.Renderer {
	var r multitemplate.Renderer
	if os.Getenv("DOCKER") == "true" {
		r = multitemplate.New()
	} else {
		r = multitemplate.NewRenderer()
	}

	// Generating our main templates
	newR, err := addTemplates(templates+"/layouts/base.tmpl", templates+"/includes/*.tmpl", efs, r)
	if err == nil {
		r = newR
	}

	// Generating our templates which do not need a header
	newR, err = addTemplates(templates+"/layouts/blank_base.tmpl", templates+"/blank_includes/*.tmpl", efs, r)
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
		r.AddFromFilesFuncsFS(filepath.Base(include), templateFuncMap, f, files...)
	}

	return r, nil
}
