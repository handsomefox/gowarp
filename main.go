package main

import (
	"fmt"
	"gowarp/pkg/keygen"
	"log"
	"net/http"
	"os"
	"strings"
)

func warp(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Server does not support Flusher!", http.StatusInternalServerError)
		return
	}

	ua := r.UserAgent()

	if strings.Contains(ua, "Firefox/") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if err := keygen.Generate(w, flusher); err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

func main() {
	http.HandleFunc("/", warp)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
