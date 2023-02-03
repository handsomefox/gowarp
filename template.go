package main

import "html/template"

const (
	basePath   = "./resources/html/"
	baseFile   = basePath + "base.html"
	footerFile = basePath + "footer.html"
)

var (
	HomeTemplate   = template.Must(template.ParseFiles(basePath+"home.html", baseFile, footerFile))
	ErrorTemplate  = template.Must(template.ParseFiles(basePath+"error.html", baseFile, footerFile))
	ConfigTemplate = template.Must(template.ParseFiles(basePath+"config.html", baseFile, footerFile))
	KeyTemplate    = template.Must(template.ParseFiles(basePath+"key.html", baseFile, footerFile))
)
