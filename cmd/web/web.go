package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/handsomefox/gowarp/pkg/server"
)

const (
	UseProxyEnvKey = "USE_PROXY"
	ConnStrEnvKey  = "DB_URI"
)

var (
	useProxy bool
	connStr  string
)

func main() {
	proxy, err := strconv.ParseBool(os.Getenv(UseProxyEnvKey))
	if err != nil {
		useProxy = false
	} else {
		useProxy = proxy
	}

	connStr = os.Getenv(ConnStrEnvKey)
	if connStr == "" {
		log.Fatalf(`No connection string was provided, exiting.
To provide a connection string, specify the enviornment variable: "%s"`, ConnStrEnvKey)
	}

	sh, err := server.NewServer(useProxy, connStr)
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           sh,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		ReadHeaderTimeout: 1 * time.Minute,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
