// config is a package for having all app configuration including keys that are used to generate new keys
package config

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	CfClientVersion = "a-6.15-2405"
	UserAgent       = "okhttp/3.12.1"
	Host            = "api.cloudflareclient.com"
	BaseURL         = "https://api.cloudflareclient.com/v0a2405"
	RingSize        = 40
	keyURL          = "https://keyses-for-generator.serdarad.repl.co/"
	// WaitTime is required to not hit rate limiting every time
	WaitTime = 45 * time.Second
)

// KeyStore represents a storage for keys, which are just strings and a mutex to update them when FetchKeys()
// is called
type KeyStore struct {
	Keys  []string
	mutex sync.Mutex
}

var (
	KeyStorage = KeyStore{
		Keys: []string{
			"7s15XT0k-90M3mz8x-J84p69Gm",
			"m42X0ct5-u3x76nH2-92bDE3F5",
			"utf2351b-6N45wfG1-396pbo0W",
			"62gz9Ro3-59pr83Gj-j472ZT8V",
			"d4IW20x1-N790rUE4-OUCT2834",
			"8n4m9bs0-X6YI357A-wH3U41s8",
			"zO1o9R76-ijtx1574-6M2X4b5v",
			"z9e0g18B-S8zx79L4-0Sf2l18o",
			"4eyK7X91-2189hWxP-39K8IzU4",
			"a6Y02n1B-5f16Ww3T-94Q1cr8p",
		},
		mutex: sync.Mutex{},
	}
)

// FetchKeys fetches the keys from the URL, parses them and updates the KeyStorage
func FetchKeys() {
	fmt.Println("Fetching the keys")
	response, err := http.Get(keyURL)
	if err != nil {
		fmt.Println("Error fetching keys")
		return
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed deconding response body")
		return
	}
	split := strings.Split(string(b), ",")

	KeyStorage.mutex.Lock()
	defer KeyStorage.mutex.Unlock()
	KeyStorage.Keys = make([]string, 0)
	KeyStorage.Keys = append(KeyStorage.Keys, split...)
	fmt.Println("Fetched the keys")
}
