package renderer

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer interface {
	echo.Renderer
}

func Template(e *echo.Echo, fsys fs.FS, root string, funcMap template.FuncMap) (TemplateRenderer, error) {
	templates := template.New("_empty")
	templates.Funcs(funcMap)

	if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			name := path[len(root):]

			// Skip trailing slash
			if name[0] == '/' {
				name = name[1:]
			}

			// Read file contents
			bytes, err := fs.ReadFile(fsys, path)
			if err != nil {
				return fmt.Errorf("failed to read file '%s': %w", path, err)
			}

			e.Logger.Infof("Compiled template: '%s' -> '%s'", path, name)

			if templates == nil {
				templates = template.Must(template.New(name).Parse(string(bytes)))
			} else {
				templates = template.Must(templates.New(name).Parse(string(bytes)))
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &templateRenderer{templates}, nil
}

// templateRenderer is a custom html/template renderer for Echo framework
type templateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if err := t.templates.ExecuteTemplate(w, name, data); err != nil {
		c.Logger().Errorf("Rendering error: %s", err)
		return err
	}

	return nil
}
