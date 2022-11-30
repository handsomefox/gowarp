package pastebin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/handsomefox/gowarp/pkg/client"
)

// ErrUnexpectedBody is returned if the format of fetched configuration was unexpected.
var ErrUnexpectedBody = errors.New("unexpected response body")

// GetClientConfiguration returns a new configuration from the hardcoded pastebin url.
func GetClientConfiguration(ctx context.Context) (*client.Configuration, error) {
	const pastebinURL = "https://pastebin.com/raw/pwtQLBiK"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pastebinURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get config: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fmt.Sprintf("error loading config from %s", pastebinURL), err)
	}
	defer res.Body.Close()

	config := &client.Configuration{}

	scanner := bufio.NewScanner(res.Body)

	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")

		if len(split) < 2 { // it should be a key=value pair
			return nil, fmt.Errorf("%w: %s", ErrUnexpectedBody, text)
		}

		key, value := split[0], split[1]

		switch key {
		case "CfClientVersion":
			config.CFClientVersion = value
		case "UserAgent":
			config.UserAgent = value
		case "Host":
			config.Host = value
		case "BaseURL":
			config.BaseURL = value
		case "Keys":
			if keys := strings.Split(value, ","); len(keys) > 0 {
				config.Keys = keys
			}
		}
	}

	return config, nil
}
