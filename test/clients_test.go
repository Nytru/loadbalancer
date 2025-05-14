package test

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

const (
	id = "192.168.97.11"
)

func TestMain(m *testing.M) {
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestAddCustomRule(t *testing.T) {
	bodyString := `{
    "id" : "` + id + `",
    "capacity" : 20,
    "refill_interval_ms": 500
	}`

	body := strings.NewReader(bodyString)

	response, err := http.Post("http://localhost:80/clients", "application/json", body)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", response.Status)
	}

	getResponse, err := http.Get("http://localhost:80/clients?id=" + id)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer getResponse.Body.Close()
	if getResponse.StatusCode != http.StatusOK {
		t.Fatalf("Expected status Not Found, got %v", getResponse.Status)
	}

	req, err := http.NewRequest(http.MethodDelete, "http://localhost:80/clients?id="+id, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	client := &http.Client{}
	response, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", response.Status)
	}

	getResponse, err = http.Get("http://localhost:80/clients?id=" + id)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer getResponse.Body.Close()
	if getResponse.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status Not Found, got %v", getResponse.Status)
	}
}

func cleanup() {
	req, err := http.NewRequest(http.MethodDelete, "http://localhost:80/clients?id="+id, nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		panic("Failed to delete client")
	}
}
