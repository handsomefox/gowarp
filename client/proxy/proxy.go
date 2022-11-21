// proxy is a package that allows to get a proxied client
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
	ErrNoProxies      = errors.New("no suitable proxy was found")
	ErrRequestFailed  = errors.New("couldn't get proxie from url")
	ErrUnexpectedBody = errors.New("unexpected proxy body")
)

type Proxy struct {
	Addr        *url.URL
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

const URL = "https://public.freeproxyapi.com/api/Proxy/Medium"

func Get(ctx context.Context) (*Proxy, error) {
	for retries := 0; retries < 15; retries++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, http.NoBody)
		if err != nil {
			return nil, ErrRequestFailed
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, ErrRequestFailed
		}
		defer resp.Body.Close() // It is fine to call defer here because the loop isn't that big.

		proxy := &Proxy{}
		if err := json.NewDecoder(resp.Body).Decode(proxy); err != nil {
			return nil, ErrUnexpectedBody
		}

		if strings.EqualFold(proxy.Type, "socks4") {
			continue
		}

		if proxy.AverageTime < 3000 {
			port := strconv.Itoa(proxy.Port)
			addr := strings.ToLower(proxy.Type) + "://" + proxy.Host + ":" + port

			URL, err := url.Parse(addr)
			if err != nil {
				continue
			}
			proxy.Addr = URL

			return proxy, nil
		}
	}
	return nil, ErrNoProxies
}
