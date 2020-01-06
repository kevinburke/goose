package goose

import (
	"os"
	"text/template"
)

// common routines

func writeTemplateToFile(path string, t *template.Template, data interface{}) (string, error) {
	f, e := os.Create(path)
	if e != nil {
		return "", e
	}

	if err := t.Execute(f, data); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	return f.Name(), nil
}
