package serdar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/handsomefox/gowarp/client/cfg"
	"github.com/handsomefox/gowarp/client/cfg/pastebin"
)

const URL = "https://keyses-for-generator.serdarad.repl.co/"

func fallback(ctx context.Context) (*cfg.Config, error) {
	fallback, err := pastebin.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the fallback config: %w", err)
	}
	return fallback, nil
}

func GetConfig(ctx context.Context) (*cfg.Config, error) {
	config, err := fallback(ctx)
	if err != nil {
		config = cfg.Default()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get config: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fmt.Sprintf("error loading config from %s", URL), err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fallback(ctx)
	}

	s := strings.Split(string(b), ",")
	if len(s) < 1 {
		return fallback(ctx)
	}
	config.Keys = s

	return config, nil
}
