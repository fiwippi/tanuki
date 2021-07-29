package main

import (
	"flag"
	"fmt"
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	inputDir := flag.String("input-dir", "", "")
	outputDir := flag.String("output-dir", "", "")
	flag.Parse()

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/js", js.Minify)

	filepath.WalkDir(*inputDir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			reader, err := os.Open(path)
			if err != nil {
				panic(err)
			}

			outputFp := *outputDir + strings.TrimPrefix(path, *inputDir)
			err = fse.EnsureFileDir(outputFp)
			if err != nil {
				panic(err)
			}

			writer, err := os.Create(outputFp)
			if err != nil {
				panic(err)
			}

			switch filepath.Ext(path) {
			case ".css":
				err = m.Minify("text/css", writer, reader)
				if err != nil {
					panic(err)
				}
			case ".js":
				err = m.Minify("text/js", writer, reader)
				if err != nil {
					panic(err)
				}
			case ".tmpl":
				err = m.Minify("text/html", writer, reader)
				if err != nil {
					panic(err)
				}
			default:
				panic(fmt.Errorf("cannot minify %s", path))
			}
		}
		return nil
	})
}
