package warp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/handsomefox/gowarp/pkg/models"
)

// Account represents a registered CF account.
type Account struct {
	ID      string `json:"id"`
	Account struct {
		License string `json:"license"`
	} `json:"account"`
	Token string `json:"token"`
}

func NewAccount(ctx context.Context, c *Client) (*Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/reg", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating a request to register account: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error registering an account: %w", err)
	}
	defer res.Body.Close()

	var acc Account
	if err := json.NewDecoder(res.Body).Decode(&acc); err != nil {
		return nil, fmt.Errorf("error decoding account data: %w", err)
	}

	return &acc, nil
}

func (acc *Account) AddReferrer(ctx context.Context, c *Client, referrer *Account) error {
	payload, err := json.Marshal(map[string]string{"referrer": referrer.ID})
	if err != nil {
		return fmt.Errorf("error marshalling account referrer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.BaseURL+"/reg/"+acc.ID, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request with account referrer: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("error adding an account referrer: %w", err)
	}
	defer res.Body.Close()

	return nil
}

func (acc *Account) RemoveDevice(ctx context.Context, c *Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+"/reg/"+acc.ID, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating a request to remove a device: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("error removing a device from account: %w", err)
	}
	defer res.Body.Close()

	return nil
}

func (acc *Account) ApplyKey(ctx context.Context, c *Client, key string) error {
	payload, err := json.Marshal(map[string]string{"license": key})
	if err != nil {
		return fmt.Errorf("error marshalling account license: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.BaseURL+"/reg/"+acc.ID+"/account",
		bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request with account key: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("error setting account's key: %w", err)
	}
	defer res.Body.Close()

	return nil
}

func (acc *Account) GetAccountData(ctx context.Context, c *Client) (*models.Account, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/reg/"+acc.ID+"/account", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request to fetch account data: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching account data: %w", err)
	}
	defer res.Body.Close()

	var accountData models.Account
	if err := json.NewDecoder(res.Body).Decode(&accountData); err != nil {
		return nil, fmt.Errorf("error decoding account data: %w", err)
	}

	return &accountData, nil
}
