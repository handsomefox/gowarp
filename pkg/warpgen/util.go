package warpgen

import (
	"crypto/tls"
	"gowarp/pkg/config"
	"net/http"
	"strings"
	"time"
)

func createClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12},
			DisableCompression:    false,
			ForceAttemptHTTP2:     false,
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func handleBrowsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	if strings.Contains(r.UserAgent(), "Firefox/") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.Header().Set("Connection", "keep-alive")
}

func setCommonHeaders(request *http.Request) {
	request.Header.Set("CF-Client-Version", config.CfClientVersion)
	request.Header.Set("Host", config.Host)
	request.Header.Set("User-Agent", config.UserAgent)
	request.Header.Set("Connection", "Keep-Alive")
}
