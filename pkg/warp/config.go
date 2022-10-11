package warp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultWaitTime = 45 * time.Second
	// keys are used for parsing a new config.
	cfClientVersionKey = "CfClientVersion"
	userAgentKey       = "UserAgent"
	hostKey            = "Host"
	baseURLKey         = "BaseURL"
	waitTimeKey        = "WaitTime"
	keysKey            = "Keys"
)

type ConfigData struct {
	CfClientVersion string
	UserAgent       string
	Host            string
	BaseURL         string
	Keys            []string
	WaitTime        time.Duration
}

type Config struct {
	cdata ConfigData
	mu    sync.Mutex
}

var ErrUnexpectedBody = errors.New("unexpected response body")

func NewConfig() *Config {
	return &Config{
		mu: sync.Mutex{},
		cdata: ConfigData{
			CfClientVersion: "a-6.15-2405",
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
			WaitTime: defaultWaitTime,
		},
	}
}

func (cfg *Config) Update(ctx context.Context, url string) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	newConfig, err := loadConfigFromURL(ctx, url)
	if err != nil {
		return fmt.Errorf("%w: %s", err, "error updating config")
	}

	cfg.cdata = newConfig.cdata
	log.Printf("%#v", cfg.cdata)

	return nil
}

func (cfg *Config) Get() ConfigData {
	cfg.mu.Lock()
	data := cfg.cdata
	cfg.mu.Unlock()
	return data
}

func loadConfigFromURL(ctx context.Context, url string) (*Config, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error creating request to get config", err)
	}

	client := newClient()
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fmt.Sprintf("error loading config from %s", url), err)
	}
	defer res.Body.Close()

	config := NewConfig()
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")

		if len(split) < 2 { // it should be a key=value pair
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedBody, text)
		}

		key, value := split[0], split[1]

		switch key {
		case cfClientVersionKey:
			config.cdata.CfClientVersion = value
		case userAgentKey:
			config.cdata.UserAgent = value
		case hostKey:
			config.cdata.Host = value
		case baseURLKey:
			config.cdata.BaseURL = value
		case waitTimeKey:
			if i, err := strconv.Atoi(value); err == nil {
				config.cdata.WaitTime = time.Duration(i) * time.Second
			}
		case keysKey:
			if keys := strings.Split(value, ","); len(keys) > 0 {
				config.cdata.Keys = keys
			}
		}
	}

	return config, nil
}
