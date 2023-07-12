package multitemplate

import (
	"embed"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin/render"
)

type Render map[string]*template.Template

func New() Render {
	return make(Render)
}

func (r Render) Add(name string, tmpl *template.Template) {
	if tmpl == nil {
		panic("template can not be nil")
	}
	if len(name) == 0 {
		panic("template name cannot be empty")
	}
	if _, ok := r[name]; ok {
		panic(fmt.Sprintf("template %s already exists", name))
	}
	r[name] = tmpl
}

func (r Render) AddFromEFS(funcMap template.FuncMap, f embed.FS, files ...string) {
	name := filepath.Base(files[len(files)-1])

	tmpl := template.New(name).Funcs(funcMap)
	for _, fp := range files {
		data, err := f.ReadFile(fp)
		if err != nil {
			panic(err)
		}
		tmpl = template.Must(tmpl.Parse(string(data)))
	}

	r.Add(name, tmpl)
}

func (r Render) Instance(name string, data interface{}) render.Render {
	return render.HTML{
		Template: r[name],
		Data:     data,
	}
}
