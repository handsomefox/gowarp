package main

import (
	"fmt"
	"gowarp/pkg/warpgen"
	"log"
	"net/http"
	"os"
)

// warp is an http.HandleFunc that generates a warp+ key and writes it to the http.ResponseWriter
func warp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := warpgen.Generate(w, r); err != nil {
		fmt.Fprintf(w, "\nError when creating keys: %v\n", err)
	}
}

func main() {
	http.HandleFunc("/", warp)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
