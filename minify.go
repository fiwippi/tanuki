package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"

	"github.com/fiwippi/tanuki/internal/fse"
)

var CSS = CSSData{
	// Light theme uses the default ColorPlaceholder
	Light: CSSTheme{
		Background:            "#FDF5E6",
		BackgroundLighter:     "#ffe2af",
		BackgroundLighterPlus: "#f3d39e",
		BackgroundUnfocus:     "#fffaf2",
		BackgroundFocus:       "white",
		Color:                 "#333",
		ColorStrong:           "black",
		Highlight:             "#ff0000",
		BorderColor:           "#f3daac",
		BorderColorDarker:     "#eac57a",
		BorderColorFocus:      "#ff0000",
		LinkColor:             "#000",
		LinkBorder:            "#999",
		LinkFocus:             "#ff0000",
		SVG:                   "#000",
		SVGHover:              "#ff0000",
	},
	// Dark theme doesn't use BorderColorDarker
	Dark: CSSTheme{
		Background:            "rgb(20, 20, 20)",
		BackgroundLighter:     "rgb(31, 31, 31)",
		BackgroundLighterPlus: "rgb(49, 49, 49)",
		BackgroundUnfocus:     "rgb(49, 49, 49)",
		BackgroundFocus:       "rgb(69, 69, 69)",
		Color:                 "rgb(245, 245, 245)",
		ColorStrong:           "white",
		ColorPlaceholder:      "rgb(217, 217, 217)",
		Highlight:             "#ff0000",
		BorderColor:           "rgb(245, 245, 245)",
		BorderColorFocus:      "#ff0000",
		LinkColor:             "rgb(245, 245, 245)",
		LinkBorder:            "#999",
		LinkFocus:             "#ff0000",
		SVG:                   "rgb(245, 245, 245)",
		SVGHover:              "#ff0000",
	},
}

type CSSData struct {
	Light, Dark CSSTheme
}

type CSSTheme struct {
	// Background
	Background            string
	BackgroundLighter     string
	BackgroundLighterPlus string
	BackgroundUnfocus     string
	BackgroundFocus       string

	// Basic colours
	Color            string
	ColorStrong      string
	ColorPlaceholder string

	// Highlighting
	Highlight string

	// Borders
	BorderColor       string
	BorderColorDarker string
	BorderColorFocus  string

	// Links
	LinkColor  string
	LinkBorder string
	LinkFocus  string

	// Images
	SVG      string
	SVGHover string
}

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
				cssText, err := ioutil.ReadAll(reader)
				if err != nil {
					panic(err)
				}

				tmpl, err := template.New("css").Parse(string(cssText))
				if err != nil {
					panic(err)
				}
				buf := bytes.NewBuffer(nil)
				err = tmpl.Execute(buf, CSS)
				if err != nil {
					panic(err)
				}

				err = m.Minify("text/css", writer, buf)
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
