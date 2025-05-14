package proxy_test

import (
	"io"
	"net/http"
	"testing"
)

func BenchmarkRoundRobin(b *testing.B) {
	url := "http://localhost:8080"
	for i := 0; i < b.N; i++ {
		response, err := http.Get(url)
		if err != nil {
			b.Fatalf("Failed to send GET request: %v", err)
		}

		resp, err := io.ReadAll(response.Body)
		if err != nil {
			b.Fatalf("Failed to read response body: %v", err)
		}
		response.Body.Close()

		stringResponse := string(resp)
		if stringResponse == "" {
			b.Fatalf("Expected non-empty response, got empty response")
		}
	}
}
