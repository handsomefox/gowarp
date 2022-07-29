package warp

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// AccountData represents the response with CF account data, including the key.
type AccountData struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

// Account represents a registered CF account.
type Account struct {
	ID      string  `json:"id"`
	Account License `json:"account"`
	Token   string  `json:"token"`
}

// License is just a license key.
type License struct {
	License string `json:"license"`
}

func (acc *Account) addReferrer(config *ConfigData, client *http.Client, second *Account) error {
	payload, _ := json.Marshal(map[string]string{
		"referrer": second.ID,
	})

	request, err := http.NewRequest("PATCH", config.BaseURL+"/reg/"+acc.ID, bytes.NewBuffer(payload))
	if err != nil {
		return ErrCreateRequest
	}

	setCommonHeaders(config, request)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc.Token)

	_, err = client.Do(request)
	if err != nil {
		return ErrAddReferrer
	}

	return nil
}

func (acc *Account) removeDevice(config *ConfigData, client *http.Client) error {
	request, err := http.NewRequest("DELETE", config.BaseURL+"/reg/"+acc.ID, http.NoBody)
	if err != nil {
		return ErrCreateRequest
	}

	setCommonHeaders(config, request)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", "Bearer "+acc.Token)

	_, err = client.Do(request)
	if err != nil {
		return ErrDeleteAccount
	}

	return nil
}

func (acc *Account) setKey(config *ConfigData, client *http.Client, key string) error {
	payload, _ := json.Marshal(map[string]string{
		"license": key,
	})
	request, err := http.NewRequest("PUT", config.BaseURL+"/reg/"+acc.ID+"/account", bytes.NewBuffer(payload))

	setCommonHeaders(config, request)

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

func (acc *Account) fetchAccountData(config *ConfigData, client *http.Client) (*AccountData, error) {
	request, err := http.NewRequest("GET", config.BaseURL+"/reg/"+acc.ID+"/account", http.NoBody)
	if err != nil {
		return nil, ErrCreateRequest
	}

	setCommonHeaders(config, request)
	request.Header.Set("Authorization", "Bearer "+acc.Token)

	response, err := client.Do(request)
	if err != nil {
		return nil, ErrFetchAccData
	}
	defer response.Body.Close()

	result := AccountData{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, ErrDecodeAccData
	}

	return &result, nil
}
