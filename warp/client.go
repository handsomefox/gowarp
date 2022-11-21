package warp

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/handsomefox/gowarp/warp/cfg"
	"github.com/handsomefox/gowarp/warp/proxy"
)

// Client is the actual client used to make requests to CF.
type Client struct {
	cl            *http.Client
	ClientVersion string
	UserAgent     string
	Host          string
	BaseURL       string
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("CF-Client-Version", c.ClientVersion)
	req.Header.Set("Host", c.Host)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Connection", "Keep-Alive")
	return c.cl.Do(req) // in this case we need the errors from the actual client, no need to wrap.
}

// Service is the service used to make requests to CF.
type Service struct {
	config *cfg.Config
	mu     sync.Mutex

	useProxy bool
}

func (c *Service) GetRequestClient(ctx context.Context) *Client {
	transport := proxiedTransport(http.ProxyFromEnvironment)
	if c.useProxy {
		px, err := proxy.Get(ctx)
		if err == nil {
			log.Printf("Using proxy: %s", px.Addr)
			transport = proxiedTransport(http.ProxyURL(px.Addr))
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

// NewService returns a *Client, if no config is specified, the DefaultConfig() is used.
func NewService(config *cfg.Config, useProxy bool) *Service {
	if config == nil {
		config = cfg.Default()
	}
	return &Service{
		config:   config,
		mu:       sync.Mutex{},
		useProxy: useProxy,
	}
}

func (c *Service) UpdateConfig(newConfig *cfg.Config) {
	log.Println("Updating config")
	log.Println(newConfig)
	c.mu.Lock()
	c.config = newConfig
	c.mu.Unlock()
}

func (c *Service) ClientVersion() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.ClientVersion
}

func (c *Service) UserAgent() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.UserAgent
}

func (c *Service) Host() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.ClientVersion
}

func (c *Service) BaseURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.BaseURL
}

func (c *Service) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	dst := make([]string, len(c.config.Keys))
	copy(dst, c.config.Keys)

	return dst
}

func (c *Service) WaitTime() time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.config.WaitTime
}

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
