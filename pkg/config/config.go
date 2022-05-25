package config

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	CfClientVersion = "a-6.13-2355"
	UserAgent       = "okhttp/3.12.1"
	Host            = "api.cloudflareclient.com"
	BaseURL         = "https://api.cloudflareclient.com/v0a2355"
	RingSize        = 20
	keyURL          = "https://keyses-for-generator.serdarad.repl.co/"
)

type KeyStore struct {
	Keys  []string
	mutex sync.Mutex
}

var (
	KeyStorage = KeyStore{
		Keys: []string{
			"10FbK2D8-VK3hZ675-h2Gx315V",
			"9D2aP6y1-ay218hX0-i0k9g8L1",
			"85Ry64SN-i8kBN514-4drF821q",
			"5Cn76eB4-r8P20v6w-9Ve0u65R",
			"0f61Z7tR-8t34mE1S-g2u846kC",
			"80cqt79O-95Y40QCm-n9u3Yt45",
			"0sMV37f4-61C5Q2Uv-82B5iYs7",
			"iH9TD408-49gsX50J-3B8fGx67",
			"1DIq085C-1g6MuL58-A8D2WM31",
			"6Tyj459c-9if2z5u3-9U3FJv52",
		},
		mutex: sync.Mutex{},
	}
)

func FetchKeys() {
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
}
