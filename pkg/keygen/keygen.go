package keygen

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gowarp/pkg/progressbar"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	client = createClient()
}

var client *http.Client

func createClient() *http.Client {
	return &http.Client{
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
	baseURL         = "https://api.cloudflareclient.com/v0a2223"
)

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

type result struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

type createdAccounts struct {
	First  *account
	Second *account
}

type account struct {
	Id      string  `json:"id"`
	Account license `json:"account"`
	Token   string  `json:"token"`
}

type license struct {
	License string `json:"license"`
}

func Generate(w http.ResponseWriter, flusher http.Flusher) error {
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

func createAccounts() (*createdAccounts, error) {
	acc1, err := register()
	if err != nil {
		return nil, err
	}

	acc2, err := register()
	if err != nil {
		return nil, err
	}

	accounts := &createdAccounts{
		First:  acc1,
		Second: acc2,
	}

	return accounts, nil
}

func register() (*account, error) {
	req, err := http.NewRequest("POST", baseURL+"/reg", nil)
	req.Header.Set("CF-Client-Version", cfClientVersion)
	req.Header.Set("User-Agent", userAgent)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var reg account
	err = json.NewDecoder(resp.Body).Decode(&reg)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &reg, err
}

func addReferrer(accounts *createdAccounts) error {
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

func deleteSecondAccount(accounts *createdAccounts) error {
	url := baseURL + fmt.Sprintf("/reg/%s", accounts.Second.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.Second.Token))
	_, err = client.Do(deleteRequest)
	return err
}

func setFirstAccountKey(accounts *createdAccounts) error {
	key := keys[rand.Intn(len(keys))]
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

func setFirstAccountLicense(accounts *createdAccounts) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"license": accounts.First.Account.License,
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

func getLicenseInformation(accounts *createdAccounts) (*result, error) {
	url := baseURL + fmt.Sprintf("/reg/%s/account", accounts.First.Id)
	getRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	getRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))

	resp, err := client.Do(getRequest)
	if err != nil {
		return nil, err
	}

	var result result
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func deleteAccount(accounts *createdAccounts) error {
	url := baseURL + fmt.Sprintf("/reg/%s", accounts.First.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accounts.First.Token))
	_, err = client.Do(deleteRequest)
	return err
}
