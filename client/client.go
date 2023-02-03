package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/handsomefox/gowarp/models"
)

type Client struct {
	cl          *http.Client
	config      *Configuration
	*log.Logger // used to log the errors and metrics.
}

func NewClient(config *Configuration, logger *log.Logger) *Client {
	if config == nil {
		config = NewConfiguration()
	}
	if logger == nil {
		logger = log.Default()
	}
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
		Logger: logger,
		config: config,
	}
}

func (c *Client) UpdateConfig(config *ConfigurationData) {
	if config == nil {
		return
	}
	c.config.Update(config)
	c.Println("Updated the config")
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("CF-Client-Version", c.config.CFClientVersion())
	req.Header.Set("Host", c.config.Host())
	req.Header.Set("User-Agent", c.config.UserAgent())
	req.Header.Set("Connection", "Keep-Alive")
	return c.cl.Do(req)
}

func (c *Client) NewAccount(ctx context.Context) (*Account, error) {
	defer func(start time.Time, logger *log.Logger) {
		logger.Println("NewAccount() took: ", time.Since(start))
	}(time.Now(), c.Logger)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL()+"/reg", http.NoBody)
	if err != nil {
		c.Println(err)
		return nil, ErrRegAccount
	}

	res, err := c.Do(req)
	if err != nil {
		c.Println(err)
		return nil, ErrRegAccount
	}
	defer res.Body.Close()

	var acc Account
	if err := json.NewDecoder(res.Body).Decode(&acc); err != nil {
		c.Println(err)
		return nil, ErrDecodeAccount
	}

	return &acc, nil
}

func (c *Client) AddReferrer(ctx context.Context, acc, referrer *Account) error {
	defer func(start time.Time, logger *log.Logger) {
		logger.Println("AddReferrer() took: ", time.Since(start))
	}(time.Now(), c.Logger)

	payload, err := json.Marshal(map[string]string{"referrer": referrer.ID})
	if err != nil {
		c.Println(err)
		return ErrEncodeAccount
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.config.BaseURL()+"/reg/"+acc.ID, bytes.NewBuffer(payload))
	if err != nil {
		c.Println(err)
		return ErrUpdateAccount
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.Println(err)
		return ErrUpdateAccount
	}
	defer res.Body.Close()

	return nil
}

func (c *Client) RemoveDevice(ctx context.Context, acc *Account) error {
	defer func(start time.Time, logger *log.Logger) {
		logger.Println("RemoveDevice() took: ", time.Since(start))
	}(time.Now(), c.Logger)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.config.BaseURL()+"/reg/"+acc.ID, http.NoBody)
	if err != nil {
		c.Println(err)
		return ErrUpdateAccount
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.Println(err)
		return ErrUpdateAccount
	}
	defer res.Body.Close()

	return nil
}

func (c *Client) ApplyKey(ctx context.Context, acc *Account, key string) error {
	defer func(start time.Time, logger *log.Logger) {
		logger.Println("ApplyKey() took: ", time.Since(start))
	}(time.Now(), c.Logger)

	payload, err := json.Marshal(map[string]string{"license": key})
	if err != nil {
		c.Println(err)
		return ErrEncodeAccount
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPut, c.config.BaseURL()+"/reg/"+acc.ID+"/account", bytes.NewBuffer(payload))
	if err != nil {
		c.Println(err)
		return ErrUpdateAccount
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.Println(err)
		return ErrUpdateAccount
	}
	defer res.Body.Close()

	return nil
}

func (c *Client) GetAccountData(ctx context.Context, acc *Account) (*models.Account, error) {
	defer func(start time.Time, logger *log.Logger) {
		logger.Println("GetAccountData() took: ", time.Since(start))
	}(time.Now(), c.Logger)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.BaseURL()+"/reg/"+acc.ID+"/account", http.NoBody)
	if err != nil {
		c.Println(err)
		return nil, ErrGetAccountData
	}

	req.Header.Set("Authorization", "Bearer "+acc.Token)

	res, err := c.Do(req)
	if err != nil {
		c.Println(err)
		return nil, ErrGetAccountData
	}
	defer res.Body.Close()

	var accountData models.Account
	if err := json.NewDecoder(res.Body).Decode(&accountData); err != nil {
		c.Println(err)
		return nil, ErrDecodeAccount
	}

	return &accountData, nil
}

// NewAccountWithLicense creates models.Account with random license.
func (c *Client) NewAccountWithLicense(ctx context.Context) (*models.Account, error) {
	defer func(start time.Time, logger *log.Logger) {
		logger.Println("NewAccountWithLicense() took: ", time.Since(start))
	}(time.Now(), c.Logger)

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

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(c.config.Keys())))) // [0; Length)
	if err != nil {
		n = big.NewInt(0)
	}

	key := c.config.Keys()[n.Int64()]
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
