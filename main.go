package main

import (
	"log"
	"net/http"
	"os"

	"gowarp/pkg/warp"
)

var warpHandle = warp.New()

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", home)
	mux.HandleFunc("/key/generate", keyGenerate)
	mux.HandleFunc("/config/update", configUpdate)

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	if err := http.ListenAndServe(":8080"+os.Getenv("PORT"), mux); err != nil {
		log.Fatal(err)
	}
}
