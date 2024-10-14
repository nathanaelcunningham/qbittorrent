package qbittorrent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestTorrentsInfo(t *testing.T) {
	// Mock a successful response for the TorrentsInfo call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/info": {
			statusCode:   http.StatusOK,
			responseBody: `[{"name": "torrent1"}, {"name": "torrent2"}]`,
		},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/v2/torrents/info"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test without parameters
	torrents, err := client.TorrentsInfo()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(torrents) != 2 {
		t.Errorf("Expected 2 torrents, got %d", len(torrents))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}

	// Test with parameters
	params := &TorrentsInfoParams{
		Filter:   "downloading",
		Category: "sample category",
		Sort:     "ratio",
	}
	expectedRequests = []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/v2/torrents/info"},
	}
	client, mockTransport, err = newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	torrents, err = client.TorrentsInfo(params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(torrents) != 2 {
		t.Errorf("Expected 2 torrents, got %d", len(torrents))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestIntegration_TorrentsInfo(t *testing.T) {
	// Set up a test server to simulate qBittorrent API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/torrents/info" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"name":"test torrent","hash":"testhash","progress":0.5}]`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := &Client{
		client:  ts.Client(),
		baseURL: ts.URL,
	}

	torrents, err := client.TorrentsInfo()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(torrents) != 1 {
		t.Errorf("Expected 1 torrent, got %d", len(torrents))
	}

	if torrents[0].Name != "test torrent" {
		t.Errorf("Expected torrent name 'test torrent', got '%s'", torrents[0].Name)
	}
}

func TestTorrentsInfo_HashesNil(t *testing.T) {
	// Mock a successful response for the TorrentsInfo call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/info": {
			statusCode:   http.StatusOK,
			responseBody: `[{"name": "torrent1"}, {"name": "torrent2"}]`,
		},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/v2/torrents/info"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with Hashes nil
	params := &TorrentsInfoParams{
		Hashes: nil,
	}
	torrents, err := client.TorrentsInfo(params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(torrents) != 2 {
		t.Errorf("Expected 2 torrents, got %d", len(torrents))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTorrentsInfo_HashesEmpty(t *testing.T) {
	// Mock a successful response for the TorrentsInfo call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/info": {
			statusCode:   http.StatusOK,
			responseBody: `[{"name": "torrent1"}, {"name": "torrent2"}]`,
		},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/v2/torrents/info"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with Hashes empty
	params := &TorrentsInfoParams{
		Hashes: []string{},
	}
	torrents, err := client.TorrentsInfo(params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(torrents) != 2 {
		t.Errorf("Expected 2 torrents, got %d", len(torrents))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTorrentsInfo_HashesSingle(t *testing.T) {
	// Mock a successful response for the TorrentsInfo call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/info": {
			statusCode:   http.StatusOK,
			responseBody: `[{"name": "torrent1"}]`,
		},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/v2/torrents/info", params: url.Values{"hashes": []string{"hash1"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with a single hash
	params := &TorrentsInfoParams{
		Hashes: []string{"hash1"},
	}
	torrents, err := client.TorrentsInfo(params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(torrents) != 1 {
		t.Errorf("Expected 1 torrent, got %d", len(torrents))
	}
	if torrents[0].Name != "torrent1" {
		t.Errorf("Expected torrent name 'torrent1', got '%s'", torrents[0].Name)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTorrentsInfo_HashesMultiple(t *testing.T) {
	// Mock a successful response for the TorrentsInfo call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/info": {
			statusCode:   http.StatusOK,
			responseBody: `[{"name": "torrent1"}, {"name": "torrent2"}]`,
		},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/v2/torrents/info", params: url.Values{"hashes": []string{"hash1|hash2"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with multiple hashes
	params := &TorrentsInfoParams{
		Hashes: []string{"hash1|hash2"},
	}
	torrents, err := client.TorrentsInfo(params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(torrents) != 2 {
		t.Errorf("Expected 2 torrents, got %d", len(torrents))
	}
	if torrents[0].Name != "torrent1" {
		t.Errorf("Expected torrent name 'torrent1', got '%s'", torrents[0].Name)
	}
	if torrents[1].Name != "torrent2" {
		t.Errorf("Expected torrent name 'torrent2', got '%s'", torrents[1].Name)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}
