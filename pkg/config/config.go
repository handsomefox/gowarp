// config is a package for having all app configuration including keys that are used to generate new keys
package config

import (
	"fmt"
	"sync"
	"time"
)

// KeyStore represents a storage for keys, which are just strings and a mutex
// to update them when FetchKeys() is called
type KeyStore struct {
	Keys  []string
	mutex sync.Mutex
}

const RingSize = 40

type ClientConfiguration struct {
	CfClientVersion string
	UserAgent       string
	Host            string
	BaseURL         string
	WaitTime        time.Duration
	KeyStorage      KeyStore
	mutex           sync.Mutex
}

var (
	ClientConfig = defaultConfig()
)

// UpdateConfig is used to update ClientConfig variable with the newest data
func UpdateConfig() {
	fmt.Println("Updating the config")
	newConfig := loadConfig()
	ClientConfig.mutex.Lock()
	defer ClientConfig.mutex.Unlock()

	ClientConfig.CfClientVersion = newConfig.CfClientVersion
	ClientConfig.UserAgent = newConfig.UserAgent
	ClientConfig.Host = newConfig.Host
	ClientConfig.BaseURL = newConfig.BaseURL
	ClientConfig.WaitTime = newConfig.WaitTime
	fmt.Printf("Updated config, current values: %v\n", ClientConfig)
}
