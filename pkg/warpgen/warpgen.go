package warpgen

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gowarp/pkg/progressbar"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	cfClientVersion = "a-6.13-2355"
	userAgent       = "okhttp/3.12.1"
	host            = "api.cloudflareclient.com"
	baseURL         = "https://api.cloudflareclient.com/v0a2355"
)

var keys = []string{
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
}

type accountData struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

type registerResponse struct {
	Id      string  `json:"id"`
	Account account `json:"account"`
	Token   string  `json:"token"`
}

type account struct {
	License string `json:"license"`
}

func Generate(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	if strings.Contains(r.UserAgent(), "Firefox/") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.Header().Set("Connection", "keep-alive")

	flusher, _ := w.(http.Flusher)
	pb := progressbar.New(w, flusher)

	client := createClient()
	pb.Update(10)
	request, err := registerRequest()
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	pb.Update(20)

	acc1 := registerResponse{}
	err = json.NewDecoder(response.Body).Decode(&acc1)
	if err != nil {
		return err
	}

	response, err = client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	pb.Update(30)

	acc2 := registerResponse{}
	err = json.NewDecoder(response.Body).Decode(&acc2)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"referrer": acc2.Id,
	})
	request, err = http.NewRequest("PATCH", baseURL+"/reg/"+acc1.Id, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	setHeaders(request)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc1.Token)

	_, err = client.Do(request)
	if err != nil {
		return err
	}
	pb.Update(40)

	request, err = http.NewRequest("DELETE", baseURL+"/reg/"+acc2.Id, nil)
	if err != nil {
		return err
	}
	setHeaders(request)
	request.Header.Set("Authorization", "Bearer "+acc2.Token)

	_, err = client.Do(request)
	if err != nil {
		return err
	}
	pb.Update(50)

	payload, _ = json.Marshal(map[string]string{
		"license": keys[rand.Intn(len(keys))],
	})
	setHeaders(request)
	request, err = http.NewRequest("PUT", baseURL+"/reg/"+acc1.Id+"/account", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc1.Token)

	_, err = client.Do(request)
	if err != nil {
		return err
	}
	pb.Update(60)

	payload, _ = json.Marshal(map[string]string{
		"license": acc1.Account.License,
	})
	request, err = http.NewRequest("PUT", baseURL+"/reg/"+acc1.Id+"/account", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	setHeaders(request)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc1.Token)

	_, err = client.Do(request)
	if err != nil {
		return err
	}
	pb.Update(70)

	request, err = http.NewRequest("GET", baseURL+"/reg/"+acc1.Id+"/account", nil)
	if err != nil {
		return err
	}
	setHeaders(request)
	request.Header.Set("Authorization", "Bearer "+acc1.Token)

	response, err = client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	pb.Update(80)

	result := accountData{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return err
	}

	request, err = http.NewRequest("DELETE", baseURL+"/reg/"+acc1.Id, nil)
	if err != nil {
		return err
	}
	setHeaders(request)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc1.Token)

	_, err = client.Do(request)
	if err != nil {
		return err
	}
	pb.Update(90)

	fmt.Fprintf(w, "\n\nAccount type: %v\nData available: %vGB\nLicense: %v\n", result.Type, result.RefCount, result.License)

	return nil
}

func setHeaders(request *http.Request) {
	request.Header.Set("CF-Client-Version", cfClientVersion)
	request.Header.Set("Host", host)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Connection", "Keep-Alive")
}

func registerRequest() (*http.Request, error) {
	request, err := http.NewRequest("POST", baseURL+"/reg", nil)
	if err != nil {
		return nil, err
	}
	setHeaders(request)

	return request, nil
}

func createClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12},
			DisableCompression:    false,
			ForceAttemptHTTP2:     false,
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
