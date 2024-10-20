package qbittorrent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock data for testing
var mockMainData = MainData{
	// Populate with appropriate fields
}

var mockTorrentPeers = TorrentPeers{
	// Populate with appropriate fields
}

func TestClient_SyncMainData(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		// Check the request URL path
		if r.URL.Path != "/api/v2/sync/maindata" {
			t.Errorf("expected path /api/v2/sync/maindata, got %s", r.URL.Path)
		}

		// Check the request parameters
		if r.URL.Query().Get("rid") != "1" {
			t.Errorf("expected rid=1, got %s", r.URL.Query().Get("rid"))
		}

		// Mock response
		w.WriteHeader(http.StatusOK)
		resp, _ := json.Marshal(mockMainData)
		w.Write(resp)
	}))
	defer mockServer.Close()

	// Create a new client with the mock server URL
	client := &Client{
		baseURL: mockServer.URL,
		client:  mockServer.Client(),
	}

	// Call the method you want to test
	result, err := client.SyncMainData(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assert the response
	if result == nil {
		t.Fatalf("expected non-nil result")
	}

	// Add more assertions based on the fields of MainData
}

func TestClient_SyncTorrentPeers(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		// Check the request URL path
		if r.URL.Path != "/api/v2/sync/torrentPeers" {
			t.Errorf("expected path /api/v2/sync/torrentPeers, got %s", r.URL.Path)
		}

		// Check the request parameters
		if r.URL.Query().Get("rid") != "1" {
			t.Errorf("expected rid=1, got %s", r.URL.Query().Get("rid"))
		}
		if r.URL.Query().Get("hash") != "somehash" {
			t.Errorf("expected hash=somehash, got %s", r.URL.Query().Get("hash"))
		}

		// Mock response
		w.WriteHeader(http.StatusOK)
		resp, _ := json.Marshal(mockTorrentPeers)
		w.Write(resp)
	}))
	defer mockServer.Close()

	// Create a new client with the mock server URL
	client := &Client{
		baseURL: mockServer.URL,
		client:  mockServer.Client(),
	}

	// Call the method you want to test
	result, err := client.SyncTorrentPeers("somehash", 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assert the response
	if result == nil {
		t.Fatalf("expected non-nil result")
	}

	// Add more assertions based on the fields of TorrentPeers
}
