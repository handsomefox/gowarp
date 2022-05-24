package warpgen

import (
	"bytes"
	"encoding/json"
	"gowarp/pkg/config"
	"net/http"
)

type accountData struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

type account struct {
	Id      string  `json:"id"`
	Account license `json:"account"`
	Token   string  `json:"token"`
}

type license struct {
	License string `json:"license"`
}

func (acc *account) addReferrer(client *http.Client, second *account) error {
	payload, _ := json.Marshal(map[string]string{
		"referrer": second.Id,
	})
	request, err := http.NewRequest("PATCH", config.BaseURL+"/reg/"+acc.Id, bytes.NewBuffer(payload))
	if err != nil {
		return ErrCreateRequest
	}
	setCommonHeaders(request)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc.Token)

	_, err = client.Do(request)
	if err != nil {
		return ErrAddReferrer
	}
	return nil
}

func (acc *account) removeDevice(client *http.Client) error {
	request, err := http.NewRequest("DELETE", config.BaseURL+"/reg/"+acc.Id, nil)
	if err != nil {
		return ErrCreateRequest
	}
	setCommonHeaders(request)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc.Token)

	_, err = client.Do(request)
	if err != nil {
		return ErrDeleteAccount
	}
	return nil
}

func (acc *account) setKey(client *http.Client, key string) error {
	payload, _ := json.Marshal(map[string]string{
		"license": key,
	})
	request, err := http.NewRequest("PUT", config.BaseURL+"/reg/"+acc.Id+"/account", bytes.NewBuffer(payload))
	setCommonHeaders(request)
	if err != nil {
		return ErrCreateRequest
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc.Token)

	_, err = client.Do(request)
	if err != nil {
		return ErrSetKey
	}
	return nil
}

func (acc *account) fetchAccountData(client *http.Client) (*accountData, error) {
	request, err := http.NewRequest("GET", config.BaseURL+"/reg/"+acc.Id+"/account", nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	setCommonHeaders(request)
	request.Header.Set("Authorization", "Bearer "+acc.Token)

	response, err := client.Do(request)
	if err != nil {
		return nil, ErrFetchAccData
	}
	defer response.Body.Close()

	result := accountData{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, ErrDecodeAccData
	}
	return &result, nil
}
