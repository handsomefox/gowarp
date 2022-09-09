package main

import (
	"html/template"
	"log"
	"net/http"
)

func init() {
	ts, err := template.ParseFiles([]string{
		"./ui/html/home.page.tmpl.html",
		"./ui/html/base.layout.tmpl.html",
		"./ui/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		panic(err)
	}

	homeTmpl = ts

	ts, err = template.ParseFiles([]string{
		"./ui/html/config.page.tmpl.html",
		"./ui/html/base.layout.tmpl.html",
		"./ui/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		panic(err)
	}

	configTmpl = ts

	ts, err = template.ParseFiles([]string{
		"./ui/html/key.page.tmpl.html",
		"./ui/html/base.layout.tmpl.html",
		"./ui/html/footer.partial.tmpl.html",
	}...)
	if err != nil {
		panic(err)
	}

	keyTmpl = ts
}

var (
	homeTmpl   *template.Template
	configTmpl *template.Template
	keyTmpl    *template.Template
)

func ErrorWithCode(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	if err := homeTmpl.Execute(w, nil); err != nil {
		log.Println(err)
		ErrorWithCode(w, http.StatusInternalServerError)
	}
}

func keyGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorWithCode(w, http.StatusMethodNotAllowed)

		return
	}

	generatedKey, err := warpHandle.GetKey()
	if err != nil {
		log.Println(err)
		ErrorWithCode(w, http.StatusInternalServerError)

		return
	}

	if err := keyTmpl.Execute(w, generatedKey); err != nil {
		log.Println(err)
		ErrorWithCode(w, http.StatusInternalServerError)
	}
}

func configUpdate(w http.ResponseWriter, r *http.Request) {
	message := "finished config update"
	if err := warpHandle.UpdateConfig(); err != nil {
		message = "failed to update config"
	}

	if err := configTmpl.Execute(w, message); err != nil {
		log.Println(err)
		ErrorWithCode(w, http.StatusInternalServerError)
	}
}
