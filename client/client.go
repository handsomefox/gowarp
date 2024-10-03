package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/handsomefox/gowarp/internal/models"
	"github.com/rs/zerolog/log"
)

type Client struct {
	cl      *http.Client
	config  *ConfigurationData
	logging bool
}

func NewClient(logging bool) *Client {
	return &Client{
		cl: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					MaxVersion: tls.VersionTLS12,
				},
				ForceAttemptHTTP2:     false,
				Proxy:                 http.ProxyFromEnvironment,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		config:  GetConfiguration(),
		logging: logging,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("CF-Client-Version", c.config.CFClientVersion)
	req.Header.Set("Host", c.config.Host)
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Connection", "Keep-Alive")
	return c.cl.Do(req)
}

func (c *Client) NewAccount(ctx context.Context) (*Account, error) {
	defer c.logTiming("NewAccount", time.Now())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/reg", http.NoBody)
	if err != nil {
		c.logError(err)
		return nil, ErrRegAccount
	}

	res, err := c.Do(req)
	if err != nil {
		c.logError(err)
		return nil, ErrRegAccount
	}
	defer res.Body.Close()

	var acc Account
	if err := json.NewDecoder(res.Body).Decode(&acc); err != nil {
		c.logError(err)
		return nil, ErrDecodeAccount
	}

	return &acc, nil
}

func (c *Client) AddReferrer(ctx context.Context, acc, referrer *Account) error {
	defer c.logTiming("AddReferrer", time.Now())

	payload, err := json.Marshal(map[string]string{"referrer": referrer.ID})
	if err != nil {
		c.logError(err)
		return ErrEncodeAccount
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.config.BaseURL+"/reg/"+acc.ID, bytes.NewBuffer(payload))
	if err != nil {
		c.logError(err)
		return ErrUpdateAccount
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.logError(err)
		return ErrUpdateAccount
	}
	defer res.Body.Close()

	return nil
}

func (c *Client) RemoveDevice(ctx context.Context, acc *Account) error {
	defer c.logTiming("RemoveDevice", time.Now())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.config.BaseURL+"/reg/"+acc.ID, http.NoBody)
	if err != nil {
		c.logError(err)
		return ErrUpdateAccount
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.logError(err)
		return ErrUpdateAccount
	}
	defer res.Body.Close()

	return nil
}

func (c *Client) ApplyKey(ctx context.Context, acc *Account, key string) error {
	defer c.logTiming("ApplyKey", time.Now())

	payload, err := json.Marshal(map[string]string{"license": key})
	if err != nil {
		c.logError(err)
		return ErrEncodeAccount
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPut, c.config.BaseURL+"/reg/"+acc.ID+"/account", bytes.NewBuffer(payload))
	if err != nil {
		c.logError(err)
		return ErrUpdateAccount
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.logError(err)
		return ErrUpdateAccount
	}
	defer res.Body.Close()

	return nil
}

func (c *Client) GetAccountData(ctx context.Context, acc *Account) (*models.Account, error) {
	defer c.logTiming("GetAccountData", time.Now())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.BaseURL+"/reg/"+acc.ID+"/account", http.NoBody)
	if err != nil {
		c.logError(err)
		return nil, ErrGetAccountData
	}

	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.logError(err)
		return nil, ErrGetAccountData
	}
	defer res.Body.Close()

	var accountData models.Account
	if err := json.NewDecoder(res.Body).Decode(&accountData); err != nil {
		c.logError(err)
		return nil, ErrDecodeAccount
	}

	return &accountData, nil
}

// NewAccountWithLicense creates models.Account with random license.
func (c *Client) NewAccountWithLicense(ctx context.Context) (*models.Account, error) {
	defer c.logTiming("NewAccountWithLicense", time.Now())

	keyAccount, err := c.NewAccount(ctx)
	if err != nil {
		return nil, err
	}

	tempAccount, err := c.NewAccount(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.AddReferrer(ctx, keyAccount, tempAccount); err != nil {
		return nil, err
	}

	if err := c.RemoveDevice(ctx, tempAccount); err != nil {
		return nil, err
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(c.config.Keys)))) // [0; Length)
	if err != nil {
		n = big.NewInt(0)
	}

	key := c.config.Keys[n.Int64()]
	if err := c.ApplyKey(ctx, keyAccount, key); err != nil {
		return nil, err
	}

	if err := c.ApplyKey(ctx, keyAccount, keyAccount.Account.License); err != nil {
		return nil, err
	}

	accountData, err := c.GetAccountData(ctx, keyAccount)
	if err != nil {
		return nil, err
	}

	if err := c.RemoveDevice(ctx, keyAccount); err != nil {
		return nil, err
	}

	return accountData, nil
}

func (c *Client) logTiming(name string, start time.Time) {
	if c.logging {
		log.Trace().Str(name+"() took", time.Since(start).String()).Send()
	}
}

func (c *Client) logError(err error) {
	if c.logging {
		log.Err(err).Send()
	}
}
