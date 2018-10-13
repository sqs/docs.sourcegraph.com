package docsite

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/sourcegraph/docsite/markdown"
)

// Generator generates a site's files.
type Generator struct {
	Sources   http.FileSystem
	Templates http.FileSystem

	AssetsURLPathPrefix string
}

func (g *Generator) getTemplate() (*template.Template, error) {
	tmpl := template.New("root")
	tmpl.Funcs(template.FuncMap{
		"asset": func(path string) string {
			return g.AssetsURLPathPrefix + path
		},
		"markdown": func(sourceFile sourceFile) template.HTML {
			return template.HTML(markdown.Run(sourceFile.Data, markdown.Options{
				Base:           &url.URL{Path: "/" + filepath.Dir(sourceFile.FilePath) + "/"},
				StripURLSuffix: ".md",
			}))
		},
	})

	// Read all template files.
	root, err := g.Templates.Open("/")
	if err != nil {
		return nil, errors.WithMessage(err, "opening templates dir")
	}
	entries, err := root.Readdir(-1)
	if err != nil {
		return nil, errors.WithMessage(err, "listing templates")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	// Parse all template files.
	//
	// TODO(sqs): support recursively listing templates
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".html" {
			continue
		}
		data, err := readFile(g.Templates, e.Name())
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("reading template %s", e.Name()))
		}
		if _, err := tmpl.Parse(string(data)); err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("parsing template %s", e.Name()))
		}
	}

	return tmpl, nil
}

func (g *Generator) getSourceFile(path string) (*sourceFile, error) {
	filePath, data, err := resolveAndReadAll(g.Sources, path)
	if err != nil {
		return nil, err
	}
	return &sourceFile{
		FilePath:    filePath,
		Data:        data,
		Breadcrumbs: makeBreadcrumbEntries(path),
	}, nil
}

func (g *Generator) Generate(path string) ([]byte, error) {
	tmpl, err := g.getTemplate()
	if err != nil {
		return nil, err
	}

	src, err := g.getSourceFile(path)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, *src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}