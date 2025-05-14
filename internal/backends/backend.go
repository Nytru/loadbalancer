package backends

import (
	"fmt"
	"net/url"
)

type Backend struct {
	Name string
	Url  url.URL
}

func NewBackend(name string, url url.URL) Backend {
	return Backend{Name: name, Url: url}
}

func parseBackendsFromConfiguration(c Configuration) ([]Backend, error) {
	backends := make([]Backend, len(c.Backends))
	for i, b := range c.Backends {
		backendUrl, err := parseURL(b.Url)
		if err != nil {
			return nil, fmt.Errorf("invalid backend \"%s\" URL: %w", b.Name, err)
		}
		backends[i] = Backend{Name: b.Name, Url: backendUrl}
	}
	return backends, nil
}

func parseURL(urlStr string) (url.URL, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return url.URL{}, fmt.Errorf("invalid URL: %s", urlStr)
	}
	return *parsedURL, nil
}
