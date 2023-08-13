package renderer

import (
	"fmt"
	"io"
	"io/fs"
	"text/template"

	"github.com/labstack/echo/v4"
)

func Template(e *echo.Echo, fsys fs.FS, root string) (echo.Renderer, error) {
	var tmpl *template.Template

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
				return fmt.Errorf("failed to read file '%s': %e", path, err)
			}

			e.Logger.Infof("Compiled template: '%s' -> '%s'", path, name)

			if tmpl == nil {
				tmpl = template.Must(template.New(name).Parse(string(bytes)))
			} else {
				tmpl = template.Must(tmpl.New(name).Parse(string(bytes)))
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &templateRenderer{tmpl}, nil
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
