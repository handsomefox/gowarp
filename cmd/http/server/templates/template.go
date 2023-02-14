package templates

import (
	"html/template"
)

type TemplateID byte

const (
	_ TemplateID = iota
	HomeID
	ErrorID
	ConfigID
	KeyID
)

type Map map[TemplateID]*template.Template

func Load() (Map, error) {
	const (
		basePath   = "./assets/html/"
		baseFile   = basePath + "base.html"
		footerFile = basePath + "footer.html"
	)
	templates := make(map[TemplateID]*template.Template)
	type TmplFile struct {
		name string
		id   TemplateID
	}
	files := []TmplFile{
		{name: "home.html", id: HomeID},
		{name: "error.html", id: ErrorID},
		{name: "config.html", id: ConfigID},
		{name: "key.html", id: KeyID},
	}
	for _, f := range files {
		tmpl, err := template.ParseFiles(basePath+f.name, baseFile, footerFile)
		if err != nil {
			return nil, err
		}
		templates[f.id] = tmpl
	}
	return templates, nil
}
