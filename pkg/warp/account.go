package warp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func NewAccount(client *http.Client, cdata *ConfigData) (*Account, error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, cdata.BaseURL+"/reg", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error creating a request to register account", err)
	}
	req = setCommonHeaders(cdata, req)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error registering an account", err)
	}
	defer res.Body.Close()

	var acc Account
	if err := json.NewDecoder(res.Body).Decode(&acc); err != nil {
		return nil, fmt.Errorf("%s: %w", "error decoding account data", err)
	}

	return &acc, nil
}

// License is just a license key.
type License struct {
	License string `json:"license"`
}

func (acc *Account) addReferrer(client *http.Client, cdata *ConfigData, other *Account) error {
	payload, err := json.Marshal(map[string]string{
		"referrer": other.ID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", "error marshalling account referrer", err)
	}

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPatch,
		cdata.BaseURL+"/reg/"+acc.ID, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("%s: %w", "error creating request with account referrer", err)
	}

	req = setCommonHeaders(cdata, req)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", "error adding an account referrer", err)
	}
	res.Body.Close()

	return nil
}

func (acc *Account) removeDevice(client *http.Client, cdata *ConfigData) error {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodDelete, cdata.BaseURL+"/reg/"+acc.ID, http.NoBody)
	if err != nil {
		return fmt.Errorf("%s: %w", "error creating a request to remove a device", err)
	}

	req = setCommonHeaders(cdata, req)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", "error removing a device from account", err)
	}
	res.Body.Close()

	return nil
}

func (acc *Account) setKey(client *http.Client, cdata *ConfigData, key string) error {
	payload, err := json.Marshal(map[string]string{
		"license": key,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", "error marshalling account license", err)
	}
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPut,
		cdata.BaseURL+"/reg/"+acc.ID+"/account", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("%s: %w", "error creating request with account key", err)
	}

	req = setCommonHeaders(cdata, req)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", "error setting account's key", err)
	}
	res.Body.Close()

	return nil
}

func (acc *Account) fetchAccountData(client *http.Client, cdata *ConfigData) (*AccountData, error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet,
		cdata.BaseURL+"/reg/"+acc.ID+"/account", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error creating request to fetch account data", err)
	}

	req = setCommonHeaders(cdata, req)
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error fetching account data", err)
	}
	defer res.Body.Close()

	var accountData AccountData
	if err := json.NewDecoder(res.Body).Decode(&accountData); err != nil {
		return nil, fmt.Errorf("%s: %w", "error decoding account data", err)
	}

	return &accountData, nil
}
