package warp

import (
	"crypto/tls"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
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

// handleBrowsers sets the Content-Type header depending on the browser User-Agent.
// for normal browser, the value is "text/event-stream", for firefox the value is "text/plain"
// this is done because if the firefox has "text/event-stream" set, instead of displaying the text,
// it tries to download a .txt file every single page update.
func handleBrowsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")

	if strings.Contains(r.UserAgent(), "Firefox/") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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

func registerAccount(client *http.Client, cdata *ConfigData) (*Account, error) {
	request, err := http.NewRequest("POST", cdata.BaseURL+"/reg", http.NoBody)
	if err != nil {
		return nil, ErrCreateRequest
	}

	request = setCommonHeaders(cdata, request)

	response, err := client.Do(request)
	if err != nil {
		return nil, ErrRegisterAccount
	}
	defer response.Body.Close()

	var acc Account
	if err := json.NewDecoder(response.Body).Decode(&acc); err != nil {
		return nil, ErrDecodeResponse
	}

	return &acc, nil
}

// randomTime is a helper function that generates a random time.Duration
// in range 0-180 seconds.
func randomTime() time.Duration {
	return time.Duration(rand.Intn(180)) * time.Second
}
