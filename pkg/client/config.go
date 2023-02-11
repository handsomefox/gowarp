package client

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Configuration wraps ConfigurationData with a mutex to allow goroutine-safe access.
type Configuration struct {
	mu   sync.RWMutex
	data *ConfigurationData
}

func NewConfiguration() *Configuration {
	return &Configuration{data: DefaultConfigurationData()}
}

// Data returns a copy of stored data.
func (cc *Configuration) Data() ConfigurationData {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return *cc.data
}

func (cc *Configuration) Update(updated *ConfigurationData) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.data = updated
}

func (cc *Configuration) CFClientVersion() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.data.CFClientVersion
}

func (cc *Configuration) UserAgent() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.data.UserAgent
}

func (cc *Configuration) Host() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.data.Host
}

func (cc *Configuration) BaseURL() string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.data.BaseURL
}

func (cc *Configuration) Keys() []string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.data.Keys
}

func (cc *Configuration) WaitTime() time.Duration {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.data.WaitTime
}

// ConfigurationData is the configuration required for the client to work.
type ConfigurationData struct {
	CFClientVersion string
	UserAgent       string
	Host            string
	BaseURL         string
	Keys            []string
	WaitTime        time.Duration
}

// DefaultConfigurationData returns usable default configuration.
func DefaultConfigurationData() *ConfigurationData {
	return &ConfigurationData{
		CFClientVersion: "a-6.15-2405",
		UserAgent:       "okhttp/3.12.1",
		Host:            "api.cloudflareclient.com",
		BaseURL:         "https://api.cloudflareclient.com/v0a2405",
		Keys: []string{
			"7s15XT0k-90M3mz8x-J84p69Gm", "m42X0ct5-u3x76nH2-92bDE3F5",
			"utf2351b-6N45wfG1-396pbo0W", "62gz9Ro3-59pr83Gj-j472ZT8V",
			"d4IW20x1-N790rUE4-OUCT2834", "8n4m9bs0-X6YI357A-wH3U41s8",
			"zO1o9R76-ijtx1574-6M2X4b5v", "z9e0g18B-S8zx79L4-0Sf2l18o",
			"4eyK7X91-2189hWxP-39K8IzU4", "a6Y02n1B-5f16Ww3T-94Q1cr8p",
		},
		WaitTime: 45 * time.Second,
	}
}

// GetConfiguration returns a new configuration parsed from the url, or an error if it fails to parse.
func GetConfiguration(ctx context.Context, url string) (*ConfigurationData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	config := &ConfigurationData{}

	scanner := bufio.NewScanner(res.Body)

	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")

		if len(split) < 2 {
			return nil, fmt.Errorf("%w: unexpected configuration format", ErrFetchingConfiguration)
		}

		key, value := split[0], split[1]

		switch key {
		case "CfClientVersion":
			config.CFClientVersion = value
		case "UserAgent":
			config.UserAgent = value
		case "Host":
			config.Host = value
		case "BaseURL":
			config.BaseURL = value
		case "Keys":
			if keys := strings.Split(value, ","); len(keys) > 0 {
				config.Keys = keys
			}
		}
	}

	if config.Keys == nil {
		return nil, fmt.Errorf("%w: no keys in fetched configuration", ErrFetchingConfiguration)
	}

	return config, nil
}
