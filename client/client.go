package client

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/handsomefox/gowarp/client/cfg"
	"github.com/handsomefox/gowarp/client/proxy"
)

// Client is the client required to make requests to work with CF's API.
type Client struct {
	mu     sync.Mutex
	client *http.Client
	config *cfg.Config
}

// NewClient returns a *Client, if no config is specified, the DefaultConfig() is used.
func NewClient(config *cfg.Config) *Client {
	if config == nil {
		config = cfg.Default()
	}
	return &Client{
		client: &http.Client{
			Transport: proxiedTransport(http.ProxyFromEnvironment),
		},
		config: config,
		mu:     sync.Mutex{},
	}
}

func (c *Client) UnuseProxy() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.client.Transport = proxiedTransport(http.ProxyFromEnvironment)
}

func (c *Client) UseProxy(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	px, err := proxy.Get(ctx)
	if err != nil {
		// use the default params
		c.client.Transport = proxiedTransport(http.ProxyFromEnvironment)
	} else {
		// use proxy
		log.Printf("Using proxy: %s", px.Addr)
		c.client.Transport = proxiedTransport(http.ProxyURL(px.Addr))
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	req.Header.Set("CF-Client-Version", c.config.ClientVersion)
	req.Header.Set("Host", c.config.Host)
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Connection", "Keep-Alive")
	return c.client.Do(req) // in this case we need the errors from the actual client, no need to wrap.
}

func (c *Client) UpdateConfig(newConfig *cfg.Config) {
	log.Println("Updating config")
	log.Println(newConfig)
	c.mu.Lock()
	c.config = newConfig
	c.mu.Unlock()
}

func (c *Client) ClientVersion() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.ClientVersion
}

func (c *Client) UserAgent() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.UserAgent
}

func (c *Client) Host() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.ClientVersion
}

func (c *Client) BaseURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config.BaseURL
}

func (c *Client) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	dst := make([]string, len(c.config.Keys))
	copy(dst, c.config.Keys)

	return dst
}

func (c *Client) WaitTime() time.Duration {
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
		DisableCompression:    false,
		ForceAttemptHTTP2:     false,
		Proxy:                 px,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
