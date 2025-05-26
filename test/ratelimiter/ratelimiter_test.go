package ratelimiter_test

import (
	"net/http"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	client := &http.Client{}
	url := "http://localhost:8080"

	for i := 0; i < 100; i++ {
		resp, err := client.Get(url)
		if err != nil {
			t.Fatalf("Failed to send GET request: %v", err)
		}
		resp.Body.Close()
	}

	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429, got %v", resp.Status)
	}
}
