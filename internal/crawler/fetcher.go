package crawler

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Fetcher retrieves the raw body for a given URL.
type Fetcher interface {
	Fetch(target string) (string, error)
}

// HTTPFetcher implements Fetcher using the net/http client.
type HTTPFetcher struct {
	Client  *http.Client
	Backoff time.Duration
}

// Fetch returns the body for both http(s) and file scheme URLs.
func (f *HTTPFetcher) Fetch(target string) (string, error) {
	if strings.HasPrefix(target, "file://") {
		parsed, err := url.Parse(target)
		if err != nil {
			return "", err
		}
		path := filepath.Clean(parsed.Path)
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
