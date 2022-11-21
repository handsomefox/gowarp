package storage

import (
	"fmt"
	"html/template"
)

type TemplateID int

const (
	Home TemplateID = iota
	Config
	Key
)

type TemplateStorage struct {
	store [3]*template.Template
}

func NewTemplateStorage() (*TemplateStorage, error) {
	ts := &TemplateStorage{
		store: [3]*template.Template{},
	}

	home, err := template.ParseFiles([]string{
		"./resources/html/home.page.tmpl.html",
		"./resources/html/base.layout.tmpl.html",
		"./resources/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "error creating home template")
	}
	ts.store[Home] = home

	config, err := template.ParseFiles([]string{
		"./resources/html/config.page.tmpl.html",
		"./resources/html/base.layout.tmpl.html",
		"./resources/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "error creating config template")
	}
	ts.store[Config] = config

	key, err := template.ParseFiles([]string{
		"./resources/html/key.page.tmpl.html",
		"./resources/html/base.layout.tmpl.html",
		"./resources/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, "error creating key template")
	}
	ts.store[Key] = key

	return ts, nil
}

func (t TemplateStorage) Home() *template.Template {
	return t.store[Home]
}

func (t TemplateStorage) Config() *template.Template {
	return t.store[Config]
}

func (t TemplateStorage) Key() *template.Template {
	return t.store[Key]
}
