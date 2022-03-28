package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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

type Result struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

type Account struct {
	Id      string  `json:"id"`
	Account license `json:"account"`
	Token   string  `json:"token"`
}

type CreatedAccounts struct {
	First  *Account
	Second *Account
}

type license struct {
	License string `json:"license"`
}

func warp(w http.ResponseWriter) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Server does not support Flusher!",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if err := doRequests(w, flusher); err != nil {
		_, _ = fmt.Fprintln(w, err)
		return
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		warp(c.Writer)
	})
	log.Fatal(r.Run())
}

func doRequests(w http.ResponseWriter, flusher http.Flusher) error {
	bar := Bar{
		Writer: &w,
	}
	bar.New(0, 100)
	UpdateProgressBar(&bar, 0, flusher)

	accounts, err := createAccounts(w)
	if err != nil {
		return err
	}
	UpdateProgressBar(&bar, 10, flusher)
	UpdateProgressBar(&bar, 20, flusher)

	if err := addReferrer(accounts); err != nil {
		return err
	}
	UpdateProgressBar(&bar, 30, flusher)

	if err := deleteSecondAccount(accounts); err != nil {
		return err
	}
	UpdateProgressBar(&bar, 40, flusher)

	key := keys[rand.Intn(len(keys))]
	UpdateProgressBar(&bar, 50, flusher)

	if err := setFirstAccountKey(key, accounts); err != nil {
		return err
	}
	UpdateProgressBar(&bar, 60, flusher)

	resp, err := getLicenseInformation(accounts)
	if err != nil {
		return err
	}
	UpdateProgressBar(&bar, 70, flusher)

	result, err := getResult(resp)
	if err != nil {
		return err
	}
	UpdateProgressBar(&bar, 80, flusher)

	if err := deleteAccount(accounts); err != nil {
		return err
	}
	UpdateProgressBar(&bar, 90, flusher)

	out := fmt.Sprintf("\n\nAccount type: %s\nData available: %sGB\nLicense: %s\n", result.Type, result.RefCount.String(), result.License)
	_, _ = fmt.Fprintln(w, out)
	return nil
}

func Register() (*Account, error) {
	req, err := http.NewRequest("POST", baseURL+"/reg", nil)
	req.Header.Add("CF-Client-Version", "a-6.3-1922")
	req.Header.Add("User-Agent", "okhttp/3.12.1")

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var reg Account
	err = toJSON(resp, &reg)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &reg, err
}

func createAccounts(w http.ResponseWriter) (*CreatedAccounts, error) {
	acc1, err := Register()
	if err != nil {
		_, _ = fmt.Fprintln(w, err)
		return nil, err
	}

	acc2, err := Register()
	if err != nil {
		_, _ = fmt.Fprintln(w, err)
		return nil, err
	}
	return &CreatedAccounts{
		First:  acc1,
		Second: acc2,
	}, nil
}

func addReferrer(accounts *CreatedAccounts) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"referrer": accounts.Second.Id,
	})
	url := baseURL + fmt.Sprintf("/reg/%s", accounts.First.Id)
	patchRequest, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	patchRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	patchRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))

	_, err = client.Do(patchRequest)
	return err
}

func deleteSecondAccount(accounts *CreatedAccounts) error {
	url := baseURL + fmt.Sprintf("/reg/%s", accounts.Second.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.Second.Token))
	_, err = client.Do(deleteRequest)
	return err
}

func setFirstAccountKey(key string, accounts *CreatedAccounts) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"license": key,
	})
	url := baseURL + fmt.Sprintf("/reg/%s/account", accounts.First.Id)
	putRequest, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))
	_, err = client.Do(putRequest)
	return err
}

func getLicenseInformation(accounts *CreatedAccounts) (*http.Response, error) {
	url := baseURL + fmt.Sprintf("/reg/%s/account", accounts.First.Id)
	getRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	getRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))
	return client.Do(getRequest)
}

func getResult(resp *http.Response) (*Result, error) {
	var result Result
	err := toJSON(resp, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func deleteAccount(accounts *CreatedAccounts) error {
	url := baseURL + fmt.Sprintf("/reg/%s", accounts.First.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))
	_, err = client.Do(deleteRequest)
	return err
}

func toJSON(r *http.Response, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
