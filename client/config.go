package client

import (
	"context"
	"os"
	"strings"
	"time"
)

// ConfigurationData is the configuration required for the client to work.
type ConfigurationData struct {
	CFClientVersion string
	UserAgent       string
	Host            string
	BaseURL         string
	Keys            []string
	WaitTime        time.Duration
}

// GetConfiguration returns a new configuration parsed from the url, or an error if it fails to parse.
func GetConfiguration(ctx context.Context) *ConfigurationData {
	return &ConfigurationData{
		CFClientVersion: os.Getenv("CFClientVersion"),
		UserAgent:       os.Getenv("UserAgent"),
		Host:            os.Getenv("Host"),
		BaseURL:         os.Getenv("BaseURL"),
		Keys:            strings.Split(os.Getenv("Keys"), ","),
		WaitTime:        45 * time.Second,
	}
}
