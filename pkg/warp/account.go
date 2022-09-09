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

func (acc *Account) addReferrer(client *http.Client, cdata *ConfigData, other *Account) error {
	payload, _ := json.Marshal(map[string]string{
		"referrer": other.ID,
	})

	req, err := http.NewRequest("PATCH", cdata.BaseURL+"/reg/"+acc.ID, bytes.NewBuffer(payload))
	if err != nil {
		return ErrCreateRequest
	}

	req = setCommonHeaders(cdata, req)

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return ErrAddReferrer
	}

	res.Body.Close()

	return nil
}

func (acc *Account) removeDevice(client *http.Client, cdata *ConfigData) error {
	req, err := http.NewRequest("DELETE", cdata.BaseURL+"/reg/"+acc.ID, http.NoBody)
	if err != nil {
		return ErrCreateRequest
	}

	req = setCommonHeaders(cdata, req)

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return ErrDeleteAccount
	}

	res.Body.Close()

	return nil
}

func (acc *Account) setKey(client *http.Client, cdata *ConfigData, key string) error {
	payload, _ := json.Marshal(map[string]string{
		"license": key,
	})
	req, err := http.NewRequest("PUT", cdata.BaseURL+"/reg/"+acc.ID+"/account", bytes.NewBuffer(payload))

	req = setCommonHeaders(cdata, req)

	if err != nil {
		return ErrCreateRequest
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return ErrSetKey
	}

	res.Body.Close()

	return nil
}

func (acc *Account) fetchAccountData(client *http.Client, cdata *ConfigData) (*AccountData, error) {
	req, err := http.NewRequest("GET", cdata.BaseURL+"/reg/"+acc.ID+"/account", http.NoBody)
	if err != nil {
		return nil, ErrCreateRequest
	}

	req = setCommonHeaders(cdata, req)

	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return nil, ErrFetchAccData
	}
	defer res.Body.Close()

	var accountData AccountData
	if err := json.NewDecoder(res.Body).Decode(&accountData); err != nil {
		return nil, ErrDecodeAccData
	}

	return &accountData, nil
}
