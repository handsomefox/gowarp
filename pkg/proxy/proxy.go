// proxy is a package that allows to get a usable proxy address from freeproxyapi.com.
package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrNoProxies        = errors.New("no suitable proxy was found")
	ErrRequestFailed    = errors.New("couldn't get proxie from url")
	ErrUnexpectedBody   = errors.New("unexpected proxy body")
	ErrFailedToParseURL = errors.New("failed to parse proxy URL")
)

// Proxy matches the response from freeproxyapi.com.
type Proxy struct {
	CountryName string  `json:"countryName"`
	Host        string  `json:"host"`
	Type        string  `json:"type"`
	ProxyLevel  string  `json:"proxyLevel"`
	Miliseconds int     `json:"miliseconds"`
	AverageTime int     `json:"averageTime"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Port        int     `json:"port"`
	IsAlive     bool    `json:"isAlive"`
}

func (p *Proxy) toURL() (*url.URL, error) {
	port := strconv.Itoa(p.Port)
	addr := strings.ToLower(p.Type) + "://" + p.Host + ":" + port

	URL, err := url.Parse(addr)
	if err != nil {
		return nil, ErrFailedToParseURL
	}

	return URL, nil
}

const apiURL = "https://public.freeproxyapi.com/api/Proxy/Medium"

// Get returns a valid URL that can be used with http.ProxyURL(URL).
func Get(ctx context.Context, retryCount int) (*url.URL, error) {
	for retries := 0; retries < retryCount; retries++ {
		proxy, err := fetchProxy(ctx)
		if err != nil {
			return nil, err
		}

		if !isUsable(proxy) {
			continue
		}

		URL, err := proxy.toURL()
		if err != nil {
			return nil, err
		}

		return URL, nil
	}

	return nil, ErrNoProxies
}

func fetchProxy(ctx context.Context) (*Proxy, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return nil, ErrRequestFailed
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrRequestFailed
	}
	defer resp.Body.Close()

	proxy := &Proxy{}
	if err := json.NewDecoder(resp.Body).Decode(proxy); err != nil {
		return nil, ErrUnexpectedBody
	}

	return proxy, nil
}

func isUsable(p *Proxy) bool {
	if strings.EqualFold(p.Type, "socks4") {
		return false
	}

	if p.AverageTime > 1500 {
		return false
	}

	return true
}
