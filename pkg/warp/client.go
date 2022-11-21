package warp

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/handsomefox/gowarp/pkg/proxy"
)

// Client is the actual client used to make requests to CF.
type Client struct {
	cl            *http.Client
	ClientVersion string
	UserAgent     string
	Host          string
	BaseURL       string
}

// Do wraps (*http.Client).Do and sets the required headers automatically.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("CF-Client-Version", c.ClientVersion)
	req.Header.Set("Host", c.Host)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Connection", "Keep-Alive")
	return c.cl.Do(req) // in this case we need the errors from the actual client, no need to wrap.
}

// Service is the service used to make requests to CF.
type Service struct {
	config *Config
	mu     sync.Mutex

	useProxy bool
}

// GetRequestClient returns a *Client configured with/without proxy that can be used in a goroutine context.
// You can only call this once, or use a new client every time since it will have a different proxy.
func (c *Service) GetRequestClient(ctx context.Context) *Client {
	transport := proxiedTransport(http.ProxyFromEnvironment)
	if c.useProxy {
		px, err := proxy.Get(ctx, 15)
		if err == nil {
			log.Printf("Using proxy: %s", px)
			transport = proxiedTransport(http.ProxyURL(px))
		}
	}

	return &Client{
		cl:            &http.Client{Transport: transport},
		ClientVersion: c.ClientVersion(),
		UserAgent:     c.UserAgent(),
		Host:          c.Host(),
		BaseURL:       c.BaseURL(),
	}
}

// NewService returns a *Server, if no config is specified, the DefaultConfig() is used.
func NewService(config *Config, useProxy bool) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	return &Service{
		config:   config,
		mu:       sync.Mutex{},
		useProxy: useProxy,
	}
}

// UpdateConfig is a thread-safe way to update underlying config to the new one.
func (c *Service) UpdateConfig(newConfig *Config) {
	c.mu.Lock()
	c.config = newConfig
	c.mu.Unlock()
}

// ClientVersion returns CF-Client-Version header from config.
func (c *Service) ClientVersion() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.ClientVersion
}

// UserAgent returns the User-Agent header from config.
func (c *Service) UserAgent() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.UserAgent
}

// Host returns the Host header from config.
func (c *Service) Host() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.ClientVersion
}

// BaseURL returns the base API URL from config.
func (c *Service) BaseURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.BaseURL
}

// Keys returns a copy of keys inside config.
func (c *Service) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	dst := make([]string, len(c.config.Keys))
	copy(dst, c.config.Keys)

	return dst
}

// WaitTime returns the sleep duration from config.
func (c *Service) WaitTime() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.config.WaitTime
}

// proxiedTransport is a shortcut to create client that works with CF's API with the specified proxy.
func proxiedTransport(px func(*http.Request) (*url.URL, error)) *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12,
		},
		ForceAttemptHTTP2:     false,
		Proxy:                 px,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
