package config

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const pastebinURL = "https://pastebin.com/raw/pwtQLBiK"

func loadConfig() *ClientConfiguration {
	response, err := http.Get(pastebinURL)
	if err != nil {
		// return default config
		return defaultConfig()
	}
	// parse the response
	defer response.Body.Close()

	config := defaultConfig()

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		// Name=Value
		text := scanner.Text()

		split := strings.Split(text, "=")
		if len(split) < 2 {
			return defaultConfig()
		}

		const (
			cfClientVersionKey = "CfClientVersion"
			userAgentKey       = "UserAgent"
			hostKey            = "Host"
			baseURLKey         = "BaseURL"
			waitTimeKey        = "WaitTime"
			keysKey            = "Keys"
		)

		switch split[0] {
		case cfClientVersionKey:
			config.CfClientVersion = split[1]
		case userAgentKey:
			config.UserAgent = split[1]
		case hostKey:
			config.Host = split[1]
		case baseURLKey:
			config.BaseURL = split[1]
		case waitTimeKey:
			i, err := strconv.Atoi(split[1])
			if err == nil {
				config.WaitTime = time.Duration(i) * time.Second
			}
		case keysKey:
			keys := strings.Split(split[1], ",")
			if len(keys) > 0 {
				config.KeyStorage.Keys = keys
			}
		}
	}
	return config
}

func defaultConfig() *ClientConfiguration {
	return &ClientConfiguration{
		CfClientVersion: "a-6.15-2405",
		UserAgent:       "okhttp/3.12.1",
		Host:            "api.cloudflareclient.com",
		BaseURL:         "https://api.cloudflareclient.com/v0a2405",
		WaitTime:        45 * time.Second,
		KeyStorage: KeyStore{
			Keys:  []string{"7s15XT0k-90M3mz8x-J84p69Gm", "m42X0ct5-u3x76nH2-92bDE3F5", "utf2351b-6N45wfG1-396pbo0W", "62gz9Ro3-59pr83Gj-j472ZT8V", "d4IW20x1-N790rUE4-OUCT2834", "8n4m9bs0-X6YI357A-wH3U41s8", "zO1o9R76-ijtx1574-6M2X4b5v", "z9e0g18B-S8zx79L4-0Sf2l18o", "4eyK7X91-2189hWxP-39K8IzU4", "a6Y02n1B-5f16Ww3T-94Q1cr8p"},
			mutex: sync.Mutex{},
		},
		mutex: sync.Mutex{},
	}
}
