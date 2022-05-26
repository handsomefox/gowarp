package main

import (
	"fmt"
	"gowarp/pkg/warpgen"
	"log"
	"net/http"
	"os"
)

// warp is a handle fucn that generates a warp+ key
func warp(w http.ResponseWriter, r *http.Request) {
	if err := warpgen.Generate(w, r); err != nil {
		fmt.Fprintf(w, "\nError when creating keys: %v\n", err)
		return
	}
}

func main() {
	http.HandleFunc("/", warp)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
