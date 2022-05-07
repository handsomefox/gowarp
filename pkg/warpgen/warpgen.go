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

var client *http.Client

func init() {
	client = &http.Client{
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
}

const (
	cfClientVersion = "a-6.11-2223"
	userAgent       = "okhttp/3.12.1"
	host            = "api.cloudflareclient.com"
	baseURL         = "https://api.cloudflareclient.com/v0a2223"
)

var headers = http.Header{
	"Host":              []string{host},
	"CF-Client-Version": []string{cfClientVersion},
	"User-Agent":        []string{userAgent},
	"Connection":        []string{"Keep-Alive"},
	"Accept-Encoding":   []string{"gzip"},
}

var keys = []string{
	"47d58Hqv-ueR37x50-db3l70n2",
	"bwK01o62-c15C78MH-Z2g5Ji74",
	"c3Dd52l4-K28SzY14-0wKO47c8",
	"Vq765xO8-TR392VI5-z4j6Xp28",
	"3C651Aqt-3PHl50V7-3751AOCb",
	"WE38q17B-25DP3um4-9A7xg48Z",
	"x0yZ894a-3Sh2b90q-718Wrw4A",
	"3z1tW5s7-03L78SFw-6w2T9F4c",
	"bkp5301R-VDB6234t-x954U8Jb",
	"ofQ759g4-9628GDiy-6i58VGa2",
}

type finalAccount struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

type registeredAccounts struct {
	First  *acc
	Second *acc
}

type acc struct {
	Id      string  `json:"id"`
	Account license `json:"account"`
	Token   string  `json:"token"`
}

type license struct {
	License string `json:"license"`
}

func Generate(w http.ResponseWriter, r *http.Request) error {
	flusher, _ := w.(http.Flusher)

	ua := r.UserAgent()
	if strings.Contains(ua, "Firefox/") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	pb := progressbar.New(w, flusher)
	pb.Update(10)

	accounts, err := createAccounts()
	if err != nil {
		return err
	}
	pb.Update(20)

	err = addReferrer(accounts)
	if err != nil {
		return err
	}
	pb.Update(30)

	err = deleteSecondAccount(accounts)
	if err != nil {
		return err
	}
	pb.Update(40)

	err = setFirstAccountKey(accounts)
	if err != nil {
		return err
	}
	pb.Update(50)

	err = setFirstAccountLicense(accounts)
	if err != nil {
		return err
	}
	pb.Update(60)

	result, err := getLicenseInformation(accounts)
	if err != nil {
		return err
	}
	pb.Update(70)

	err = deleteAccount(accounts)
	if err != nil {
		return err
	}
	pb.Update(80)

	out := fmt.Sprintf("\n\nAccount type: %v\nData available: %vGB\nLicense: %v\n", result.Type, result.RefCount, result.License)
	pb.Update(90)

	fmt.Fprintln(w, out)
	return nil
}

func createAccounts() (*registeredAccounts, error) {
	first, err := register()
	if err != nil {
		return nil, err
	}

	second, err := register()
	if err != nil {
		return nil, err
	}

	accounts := &registeredAccounts{
		First:  first,
		Second: second,
	}

	return accounts, nil
}

func register() (*acc, error) {
	request, err := http.NewRequest("POST", baseURL+"/reg", nil)
	request.Header = headers

	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	var registered acc
	err = json.NewDecoder(response.Body).Decode(&registered)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &registered, err
}

func addReferrer(accounts *registeredAccounts) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"referrer": accounts.Second.Id,
	})

	url := baseURL + fmt.Sprintf("/reg/%s", accounts.First.Id)
	request, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	request.Header = headers
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))

	_, err = client.Do(request)
	return err
}

func deleteSecondAccount(accounts *registeredAccounts) error {
	url := baseURL + fmt.Sprintf("/reg/%s", accounts.Second.Id)
	request, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		return err
	}

	request.Header = headers
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.Second.Token))

	_, err = client.Do(request)
	return err
}

func setFirstAccountKey(accounts *registeredAccounts) error {
	key := keys[rand.Intn(len(keys))]

	payload, _ := json.Marshal(map[string]interface{}{
		"license": key,
	})

	url := baseURL + fmt.Sprintf("/reg/%s/account", accounts.First.Id)
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	request.Header = headers
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))

	_, err = client.Do(request)
	return err
}

func setFirstAccountLicense(accounts *registeredAccounts) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"license": accounts.First.Account.License,
	})

	url := baseURL + fmt.Sprintf("/reg/%s/account", accounts.First.Id)
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header = headers
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))

	_, err = client.Do(request)
	return err
}

func getLicenseInformation(accounts *registeredAccounts) (*finalAccount, error) {
	url := baseURL + fmt.Sprintf("/reg/%s/account", accounts.First.Id)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header = headers
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	var result finalAccount
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func deleteAccount(accounts *registeredAccounts) error {
	url := baseURL + fmt.Sprintf("/reg/%s", accounts.First.Id)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	request.Header = headers
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))

	_, err = client.Do(request)
	return err
}
