package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Response struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

type Account struct {
	License string `json:"license"`
}

type Registered struct {
	Id      string  `json:"id"`
	Account Account `json:"account"`
	Token   string  `json:"token"`
}

var keys = []string{
	"7f625kCI-cB450gH1-G586eSA3",
	"3791MrNh-f5736byE-u12i6zW3",
	"m07O3Zf2-Z69CB71z-547fows8",
	"6a80lV4c-aj39v02s-6mJp2b54",
	"6Sh1X0W2-W61F9Da4-6Jw02cA3",
	"9tZJ0W58-62a93xUu-I926xq1P",
	"p07riY35-92j35Mdb-Dt857b2P",
	"E6Oq5h01-1Vc0x4d9-u4v0J13E",
	"1R2wl6c0-t16p5qm4-rh01E73L",
	"78H9Mfk2-1u7N5k4T-nq50K1M8",
}

const baseURL = "https://api.cloudflareclient.com/v0a1922"

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:     false,
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func warp(w http.ResponseWriter) {
	doWork(w)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		warp(c.Writer)
	})
	log.Fatal(r.Run())
}

func doWork(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Server does not support Flusher!",
			http.StatusInternalServerError)
		return
	}

	bar := Bar{
		Writer: &w,
	}
	bar.New(0, 100)
	bar.Play(0)

	acc1, err := Register()
	UpdateProgressBar(&bar, 0, flusher)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	acc2, err := Register()
	UpdateProgressBar(&bar, 10, flusher)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"referrer": acc2.Id,
	})
	url := baseURL + fmt.Sprintf("/reg/%s", acc1.Id)
	patchRequest, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	patchRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	patchRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc1.Token))

	_, err = client.Do(patchRequest)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 20, flusher)

	url = baseURL + fmt.Sprintf("/reg/%s", acc2.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc2.Token))
	_, err = client.Do(deleteRequest)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 30, flusher)

	key := keys[rand.Intn(len(keys))]
	payload, _ = json.Marshal(map[string]interface{}{
		"license": key,
	})
	url = baseURL + fmt.Sprintf("/reg/%s/account", acc1.Id)
	putRequest, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc1.Token))
	_, err = client.Do(putRequest)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 40, flusher)

	payload, _ = json.Marshal(map[string]interface{}{
		"license": key,
	})
	url = baseURL + fmt.Sprintf("/reg/%s/account", acc1.Id)
	putRequest, err = http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc1.Token))
	_, err = client.Do(putRequest)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 50, flusher)

	payload, _ = json.Marshal(map[string]interface{}{
		"license": acc1.Account.License,
	})
	url = baseURL + fmt.Sprintf("/reg/%s/account", acc1.Id)
	putRequest, err = http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc1.Token))
	_, err = client.Do(putRequest)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 60, flusher)

	url = baseURL + fmt.Sprintf("/reg/%s/account", acc1.Id)
	getRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	getRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc1.Token))
	resp, err := client.Do(getRequest)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 70, flusher)

	var result Response
	err = toJSON(resp, &result)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 80, flusher)

	url = baseURL + fmt.Sprintf("/reg/%s", acc1.Id)
	deleteRequest, err = http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc1.Token))
	_, err = client.Do(deleteRequest)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	UpdateProgressBar(&bar, 90, flusher)

	out := fmt.Sprintf("\n\nAccount type: %s\nData available: %sGB\nLicense: %s\n", result.Type, result.RefCount.String(), result.License)
	fmt.Fprintln(w, out)
}

func Register() (Registered, error) {
	req, err := http.NewRequest("POST", baseURL+"/reg", nil)

	req.Header.Add("CF-Client-Version", "a-6.3-1922")
	req.Header.Add("User-Agent", "okhttp/3.12.1")

	if err != nil {
		fmt.Println(err)
		return Registered{}, err
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return Registered{}, err
	}

	var reg Registered
	err = toJSON(resp, &reg)

	if err != nil {
		fmt.Println(err)
		return Registered{}, err
	}

	return reg, err
}

func toJSON(r *http.Response, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
