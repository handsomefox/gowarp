package warp

import (
	"crypto/tls"
	"net/http"
	"time"
)

// newClient returns a pointer to http.Client which is set up to work with cloudflare APIs
// if normal client is used, cloudflare returns an HTTP403 response.
func newClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
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

// setCommonHeaders is a helper function that sets headers required for each request to cloudflare APIs.
func setCommonHeaders(cdata *ConfigData, r *http.Request) *http.Request {
	r.Header.Set("CF-Client-Version", cdata.CfClientVersion)
	r.Header.Set("Host", cdata.Host)
	r.Header.Set("User-Agent", cdata.UserAgent)
	r.Header.Set("Connection", "Keep-Alive")
	return r
}
