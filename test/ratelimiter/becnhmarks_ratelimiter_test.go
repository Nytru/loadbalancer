package ratelimiter_test

import (
	"net/http"
	"testing"
)

func BenchmarkRateLimiter(b *testing.B) {
	client := &http.Client{}
	url := "http://localhost:8080"

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(url)
		if err != nil {
			b.Fatalf("Failed to send GET request: %v", err)
		}
		resp.Body.Close()
	}
}
