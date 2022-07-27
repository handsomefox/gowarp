package main

import (
	"log"
	"net/http"
	"os"

	"gowarp/warp"
)

func main() {
	mux := http.NewServeMux()

	warpHandle := warp.New()

	mux.Handle("/", warpHandle)
	mux.Handle("/config/update", warpHandle.GetConfig())

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
}
