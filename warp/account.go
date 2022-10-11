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

func NewAccount(ctx context.Context, cdata *ConfigData) (*Account, error) {
	client := newClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cdata.BaseURL+"/reg", http.NoBody)
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

func (acc *Account) addReferrer(ctx context.Context, cfg *ConfigData, other *Account) error {
	client := newClient()
	payload, err := json.Marshal(map[string]string{
		"referrer": other.ID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", "error marshalling account referrer", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, cfg.BaseURL+"/reg/"+acc.ID, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("%s: %w", "error creating request with account referrer", err)
	}

	req = setCommonHeaders(cfg, req)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", "error adding an account referrer", err)
	}
	res.Body.Close()

	return nil
}

func (acc *Account) removeDevice(ctx context.Context, cfg *ConfigData) error {
	client := newClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, cfg.BaseURL+"/reg/"+acc.ID, http.NoBody)
	if err != nil {
		return fmt.Errorf("%s: %w", "error creating a request to remove a device", err)
	}

	req = setCommonHeaders(cfg, req)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", "error removing a device from account", err)
	}
	res.Body.Close()

	return nil
}

func (acc *Account) setKey(ctx context.Context, cfg *ConfigData, key string) error {
	client := newClient()
	payload, err := json.Marshal(map[string]string{
		"license": key,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", "error marshalling account license", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, cfg.BaseURL+"/reg/"+acc.ID+"/account",
		bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("%s: %w", "error creating request with account key", err)
	}

	req = setCommonHeaders(cfg, req)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", "error setting account's key", err)
	}
	res.Body.Close()

	return nil
}

func (acc *Account) fetchAccountData(ctx context.Context, cfg *ConfigData) (*AccountData, error) {
	client := newClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.BaseURL+"/reg/"+acc.ID+"/account", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "error creating request to fetch account data", err)
	}

	req = setCommonHeaders(cfg, req)
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
