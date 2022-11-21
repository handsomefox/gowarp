package warp

import (
	"time"
)

// Config is the configuration required for the client to work.
type Config struct {
	ClientVersion string
	UserAgent     string
	Host          string
	BaseURL       string
	Keys          []string
	WaitTime      time.Duration
}

// DefaultConfig returns usable default configuration.
func DefaultConfig() *Config {
	return &Config{
		ClientVersion: "a-6.15-2405",
		UserAgent:     "okhttp/3.12.1",
		Host:          "api.cloudflareclient.com",
		BaseURL:       "https://api.cloudflareclient.com/v0a2405",
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
