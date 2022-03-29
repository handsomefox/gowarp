package keygen

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gowarp/internal/progressbar"
	"math/rand"
	"net/http"
	"time"
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

func Generate(w http.ResponseWriter, flusher http.Flusher) error {
	pb := progressbar.Progressbar{
		Writer:  w,
		Flusher: flusher,
	}
	pb.Init(0, 100)
	pb.Update(0)

	kg := keygen{
		client: createClient(),
		writer: w,
	}
	pb.Update(10)

	if err := kg.createAccounts(); err != nil {
		return err
	}
	pb.Update(20)

	if err := kg.addReferrer(); err != nil {
		return err
	}
	pb.Update(30)

	if err := kg.deleteSecondAccount(); err != nil {
		return err
	}
	pb.Update(40)

	if err := kg.setFirstAccountKey(); err != nil {
		return err
	}
	pb.Update(50)

	if err := kg.setFirstAccountLicense(); err != nil {
		return err
	}
	pb.Update(60)

	result, err := kg.getLicenseInformation()
	if err != nil {
		return err
	}
	pb.Update(70)

	if err := kg.deleteAccount(); err != nil {
		return err
	}
	pb.Update(80)

	out := fmt.Sprintf("\n\nAccount type: %s\nData available: %sGB\nLicense: %s\n", result.Type, result.RefCount.String(), result.License)
	pb.Update(90)

	_, _ = fmt.Fprintln(w, out)
	return nil
}

type keygen struct {
	accounts *createdAccounts
	client   *http.Client
	writer   http.ResponseWriter
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

func (kg *keygen) createAccounts() error {
	acc1, err := kg.register()
	if err != nil {
		return err
	}

	acc2, err := kg.register()
	if err != nil {
		return err
	}

	kg.accounts = &createdAccounts{
		First:  acc1,
		Second: acc2,
	}

	return nil
}

func (kg *keygen) register() (*account, error) {
	req, err := http.NewRequest("POST", baseURL+"/reg", nil)
	req.Header.Add("CF-Client-Version", "a-6.3-1922")
	req.Header.Add("User-Agent", "okhttp/3.12.1")

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp, err := kg.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var reg account
	err = toJSON(resp, &reg)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &reg, err
}

func (kg *keygen) addReferrer() error {
	payload, _ := json.Marshal(map[string]interface{}{
		"referrer": kg.accounts.Second.Id,
	})
	url := baseURL + fmt.Sprintf("/reg/%s", kg.accounts.First.Id)
	patchRequest, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	patchRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	patchRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))

	_, err = kg.client.Do(patchRequest)
	return err
}

func (kg *keygen) deleteSecondAccount() error {
	url := baseURL + fmt.Sprintf("/reg/%s", kg.accounts.Second.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.Second.Token))
	_, err = kg.client.Do(deleteRequest)
	return err
}

func (kg *keygen) setFirstAccountKey() error {
	key := keys[rand.Intn(len(keys))]
	payload, _ := json.Marshal(map[string]interface{}{
		"license": key,
	})
	url := baseURL + fmt.Sprintf("/reg/%s/account", kg.accounts.First.Id)
	putRequest, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))
	_, err = kg.client.Do(putRequest)
	return err
}

func (kg *keygen) setFirstAccountLicense() error {
	payload, _ := json.Marshal(map[string]interface{}{
		"license": kg.accounts.First.Account.License,
	})
	url := baseURL + fmt.Sprintf("/reg/%s/account", kg.accounts.First.Id)
	putRequest, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))
	_, err = kg.client.Do(putRequest)
	return err
}

func (kg *keygen) getLicenseInformation() (*result, error) {
	url := baseURL + fmt.Sprintf("/reg/%s/account", kg.accounts.First.Id)
	getRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	getRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))

	resp, err := kg.client.Do(getRequest)
	if err != nil {
		return nil, err
	}

	var result result
	if err := toJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (kg *keygen) deleteAccount() error {
	url := baseURL + fmt.Sprintf("/reg/%s", kg.accounts.First.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))
	_, err = kg.client.Do(deleteRequest)
	return err
}

func toJSON(r *http.Response, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
