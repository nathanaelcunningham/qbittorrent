package qbittorrent

import (
	"net/http"
	"testing"
)

func TestTorrentsExport(t *testing.T) {
	expectedData := "torrent file data"
	// Mock successful AuthLogin and TorrentsExport responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":      {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/export": {statusCode: http.StatusOK, responseBody: expectedData},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "POST", url: "/api/v2/torrents/export"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data, err := client.TorrentsExport("testhash")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != expectedData {
		t.Errorf("Expected %s, got %s", expectedData, string(data))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTorrentsAdd(t *testing.T) {
	// Mock successful AuthLogin and TorrentsAdd responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":   {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/add": {statusCode: http.StatusOK, responseBody: "Ok."},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "POST", url: "/api/v2/torrents/add"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.TorrentsAdd("test.torrent", []byte("torrent data"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTorrentsDelete(t *testing.T) {
	// Mock successful AuthLogin and TorrentsDelete responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":      {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/delete": {statusCode: http.StatusOK, responseBody: "Ok."},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "POST", url: "/api/v2/torrents/delete"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.TorrentsDelete("testhash")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestSetForceStart(t *testing.T) {
	// Mock successful AuthLogin and SetForceStart responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":             {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/setForceStart": {statusCode: http.StatusOK, responseBody: "Ok."},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "POST", url: "/api/v2/torrents/setForceStart"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.SetForceStart("testhash", true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestTorrentsTrackers(t *testing.T) {
	responseBody := `[{"url":"tracker1","status":1},{"url":"tracker2","status":0}]`
	// Mock successful AuthLogin and TorrentsTrackers responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":        {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/trackers": {statusCode: http.StatusOK, responseBody: responseBody},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/v2/torrents/trackers"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	trackers, err := client.TorrentsTrackers("testhash")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(trackers) != 2 {
		t.Errorf("Expected 2 trackers, got %d", len(trackers))
	}

	if trackers[0].URL != "tracker1" {
		t.Errorf("Expected tracker URL 'tracker1', got '%s'", trackers[0].URL)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}
