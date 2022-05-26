package warpgen

import (
	"crypto/tls"
	"gowarp/pkg/config"
	"net/http"
	"strings"
	"time"
)

// createClient returns a pointer to http.Client which is set up to work with cloudflare APIs
// if normal client is used, cloudflare returns an HTTP403 response.
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

// handleBrowsers sets the Content-Type header depending on the browser User-Agent.
// for normal browser, the value is "text/event-stream", for firefox the value is "text/plain"
// this is done because if the firefox has "text/event-stream" set, instead of displaing the text,
// it tries to download a .txt file every single page update.
func handleBrowsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	if strings.Contains(r.UserAgent(), "Firefox/") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.Header().Set("Connection", "keep-alive")
}

// setCommonHeaders is a helper function that sets headers required for each request to cloudflare APIs
func setCommonHeaders(request *http.Request) {
	request.Header.Set("CF-Client-Version", config.ClientConfig.CfClientVersion)
	request.Header.Set("Host", config.ClientConfig.Host)
	request.Header.Set("User-Agent", config.ClientConfig.UserAgent)
	request.Header.Set("Connection", "Keep-Alive")
}
