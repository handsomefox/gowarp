package warp

import (
	"bufio"
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
	data ConfigData
	mu   sync.Mutex
}

func defaultConfig() *Config {
	return &Config{
		mu: sync.Mutex{},
		data: ConfigData{
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

func NewConfig() *Config {
	return defaultConfig()
}

// triggers updates when requested.
func (c *Config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := c.Update(pastebinURL); err != nil { // FIXME: hardcoded URL
		fmt.Fprintln(w, fmt.Errorf("error updating config: %w", err))
	}

	fmt.Fprintln(w, "finished config update")
}

func (c *Config) Update(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	newConfig, err := loadConfigFromURL(url)
	if err != nil {
		return fmt.Errorf("error updating config")
	}

	c.data = newConfig.data

	log.Printf("%#v", c.data)

	return nil
}

func (c *Config) Get() ConfigData {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.data
}

func loadConfigFromURL(url string) (*Config, error) {
	response, err := http.Get(url) //nolint:gosec // the url is a code constant, not user input
	if err != nil {
		return nil, fmt.Errorf("error %w when loading config from %v", err, url)
	}
	defer response.Body.Close()

	config := defaultConfig()

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		text := scanner.Text()

		split := strings.Split(text, "=")
		if len(split) < 2 {
			return nil, fmt.Errorf("unexpected config body: %v", text)
		}

		switch split[0] {
		case cfClientVersionKey:
			config.data.CfClientVersion = split[1]
		case userAgentKey:
			config.data.UserAgent = split[1]
		case hostKey:
			config.data.Host = split[1]
		case baseURLKey:
			config.data.BaseURL = split[1]
		case waitTimeKey:
			i, err := strconv.Atoi(split[1])
			if err == nil {
				config.data.WaitTime = time.Duration(i) * time.Second
			}
		case keysKey:
			keys := strings.Split(split[1], ",")
			if len(keys) > 0 {
				config.data.Keys = keys
			}
		}
	}

	return config, nil
}
