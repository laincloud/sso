/*
Package render provides html/templates rendering for the composable templates which can be called a set of templates.

	// we define some layout/desktop.html
	{{ template "header" .}}
	{{ template "content" .}}

	// then in the common/header.html, we define
	{{ define "header" }} ... {{ end }}

	// then we define content inside some index.html
	{{ define "content" }}<p>Hello {{ .Hello }}</p>{{ end }}

	// we can link those fragments as a render.TemplateSet
	ts := render.TemplateSet("index", "desktop.html", "common/header.html", "index.html", "layout/desktop.html")

	// New a renderer with some options
	r := render.New(render.Options{
		Directory:     "templates",
		Funcs:         []template.FuncMap{someServer.DefaultRouteFuncs()},
		Delims:        render.Delims{"{{", "}}"},
		IndentJson:    true,
		UseBufPool:    true,
		IsDevelopment: false,
	}, []*render.TemplateSet{ts})

	// e.g. render some named templates
	err := r.Html(w, http.StatusOK, "index", map[string]string{ "Hello": "World" })

	// or we can render some json data
	err := r.Json(w, http.StatusOK, "Hello World")

*/
package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"

	"github.com/mijia/sweb/log"
	"github.com/oxtoacart/bpool"
)

const (
	kContentCharset = "; charset=UTF-8"
	kContentHtml    = "text/html"
	kContentJson    = "application/json"
)

// Delims defines the deliminator for html/template.
type Delims struct {
	Left  string
	Right string
}

// Options defines basic options for the renderer.
type Options struct {
	// Templates lookup directory
	Directory string

	// User defined functions passed to the template engine
	Funcs []template.FuncMap

	// Deliminator for the html/template
	Delims Delims

	// Should indent json for readable format when call r.Json
	IndentJson bool

	// Should use the buf pool for the rendering or just use the bytes.Buffer
	UseBufPool bool

	// In development mode, the templates would be recompiled for each rendering call
	IsDevelopment bool
}

// TemplateSet defines a template set for composable template fragments
type TemplateSet struct {
	name     string
	fileList []string
	entry    string
	template *template.Template
}

// NewTemplateSet returns a new template set, name provides the name for reverse searching, entry defines the most top
// template name, like the "layout.html"; have to at least give a template file or multiple template files used in this set.
func NewTemplateSet(name string, entry string, tFile string, otherFiles ...string) *TemplateSet {
	fileList := make([]string, 0, 1+len(otherFiles))
	fileList = append(fileList, tFile)
	for _, f := range otherFiles {
		fileList = append(fileList, f)
	}
	return &TemplateSet{
		name:     name,
		fileList: fileList,
		entry:    entry,
	}
}

// Render defines basic renderer data.
type Render struct {
	opt       Options
	templates map[string]*TemplateSet
	bufPool   *bpool.BufferPool
}

// Json renders a object and writes the result and status to http.ResponseWriter
func (r *Render) Json(w http.ResponseWriter, status int, v interface{}) error {
	var (
		data []byte
		err  error
	)
	if r.opt.IndentJson {
		data, err = json.MarshalIndent(v, "", "  ")
		data = append(data, '\n')
	} else {
		data, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", kContentJson+kContentCharset)
	w.WriteHeader(status)
	_, err = w.Write(data)
	return err
}

// Html renders a named template set with given data binding, and writes the result and status to http.ResponseWriter
func (r *Render) Html(w http.ResponseWriter, status int, name string, binding interface{}) error {
	if r.opt.IsDevelopment {
		r.compile()
	}

	if tSet, ok := r.templates[name]; !ok {
		return fmt.Errorf("Cannot find template %q in Render", name)
	} else {
		if r.opt.UseBufPool {
			buf := r.bufPool.Get()
			if err := tSet.template.Execute(buf, binding); err != nil {
				return fmt.Errorf("Template execution error, %s", err)
			}
			w.Header().Set("Content-Type", kContentHtml+kContentCharset)
			w.WriteHeader(status)
			_, err := buf.WriteTo(w)
			r.bufPool.Put(buf)
			return err
		} else {
			out := new(bytes.Buffer)
			if err := tSet.template.Execute(out, binding); err != nil {
				return fmt.Errorf("Template execution error, %s", err)
			}
			w.Header().Set("Content-Type", kContentHtml+kContentCharset)
			w.WriteHeader(status)
			_, err := io.Copy(w, out)
			return err
		}
	}
}

// compile all the templates
func (r *Render) compile() {
	for _, ts := range r.templates {
		fileList := make([]string, len(ts.fileList))
		for i, f := range ts.fileList {
			fileList[i] = path.Join(r.opt.Directory, f)
		}
		ts.template = template.New(ts.entry)
		ts.template.Delims(r.opt.Delims.Left, r.opt.Delims.Right)
		for _, funcs := range r.opt.Funcs {
			ts.template.Funcs(funcs)
		}
		ts.template = template.Must(ts.template.ParseFiles(fileList...))
	}
	log.Debugf("Templates have been compiled, count=%d", len(r.templates))
}

// New a renderer with given template sets and options.
func New(opt Options, tSets []*TemplateSet) *Render {
	r := &Render{
		opt:       opt,
		templates: make(map[string]*TemplateSet),
	}
	if opt.UseBufPool {
		r.bufPool = bpool.NewBufferPool(64)
	}
	for _, ts := range tSets {
		r.templates[ts.name] = ts
	}
	r.compile()
	return r
}
