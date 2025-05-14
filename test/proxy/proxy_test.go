package proxy_test

import (
	"io"
	"net/http"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	url := "http://localhost:8080"
	response1, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer response1.Body.Close()

	response2, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer response2.Body.Close()

	resp1, err := io.ReadAll(response1.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	resp2, err := io.ReadAll(response2.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	stringResponse1 := string(resp1)
	stringResponse2 := string(resp2)
	if stringResponse1 == stringResponse2 {
		t.Fatalf("Expected different responses, got same response: %s", string(resp1))
	}
}
