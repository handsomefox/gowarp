// package pastebin allows fetching a client config from pastebin.
package pastebin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/handsomefox/gowarp/pkg/warp"
)

// ErrUnexpectedBody is returned if the format of fetched configuration was unexpected.
var ErrUnexpectedBody = errors.New("unexpected response body")

const pastebinURL = "https://pastebin.com/raw/pwtQLBiK"

const (
	cfClientVersionKey = "CfClientVersion"
	userAgentKey       = "UserAgent"
	hostKey            = "Host"
	baseURLKey         = "BaseURL"
	waitTimeKey        = "WaitTime"
	keysKey            = "Keys"
)

// GetConfig returns a new configuration from the hardcoded pastebin url.
func GetConfig(ctx context.Context) (*warp.Config, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pastebinURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get config: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fmt.Sprintf("error loading config from %s", pastebinURL), err)
	}
	defer res.Body.Close()

	config := &warp.Config{
		ClientVersion: cfClientVersionKey,
		UserAgent:     userAgentKey,
		Host:          hostKey,
		BaseURL:       baseURLKey,
		Keys:          []string{},
		WaitTime:      45 * time.Second,
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")

		if len(split) < 2 { // it should be a key=value pair
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedBody, text)
		}

		key, value := split[0], split[1]

		switch key {
		case cfClientVersionKey:
			config.ClientVersion = value
		case userAgentKey:
			config.UserAgent = value
		case hostKey:
			config.Host = value
		case baseURLKey:
			config.BaseURL = value
		case waitTimeKey:
			if i, err := strconv.Atoi(value); err == nil {
				config.WaitTime = time.Duration(i) * time.Second
			}
		case keysKey:
			if keys := strings.Split(value, ","); len(keys) > 0 {
				config.Keys = keys
			}
		}
	}

	return config, nil
}
