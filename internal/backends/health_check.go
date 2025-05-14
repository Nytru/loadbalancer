package backends

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type healthCheck struct {
	Timeout  time.Duration
	Interval time.Duration
	Path     url.URL
	Method   string
}

func parseHealthCheck(c Configuration) (healthCheck, error) {
	if c.HealthCheck.Url == "" {
		c.HealthCheck.Url = "/health"
	}
	if c.HealthCheck.Method == "" {
		c.HealthCheck.Method = http.MethodGet
	}
	if c.HealthCheck.Timeout == 0 {
		c.HealthCheck.Timeout = 10 * time.Second
	}
	if c.HealthCheck.Interval == 0 {
		c.HealthCheck.Interval = 30 * time.Second
	}

	healthCheckURL := url.URL{Path: c.HealthCheck.Url}
	return healthCheck{
		Timeout:  c.HealthCheck.Timeout,
		Interval: c.HealthCheck.Interval,
		Path:     healthCheckURL,
		Method:   c.HealthCheck.Method,
	}, nil
}
func (hc *healthCheck) checkBackend(ctx context.Context, backend Backend, httpClient *http.Client) (bool, error) {
	fullURL := backend.Url.ResolveReference(&hc.Path).String()

	ctx, cancel := context.WithTimeout(ctx, hc.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, hc.Method, fullURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false, nil // Not an error, just not alive
	}
	defer resp.Body.Close()
	return true, nil
}
