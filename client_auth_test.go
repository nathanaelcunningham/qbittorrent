package qbittorrent

import (
	"net/http"
	"testing"
)

func TestNewClientWithoutAuth(t *testing.T) {
	// Test without authentication using a mock client
	endpointResponses := map[string]mockResponse{}
	expectedRequests := []expectedRequest{}

	mockTransport := &mockRoundTripper{
		endpointResponses: endpointResponses,
		expectedRequests:  expectedRequests,
		requestIndex:      0,
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client, err := NewClient("", "", "localhost", "8080", httpClient)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.username != "" || client.password != "" {
		t.Errorf("Expected empty username and password")
	}
}

func TestNewClientWithAuth(t *testing.T) {
	// Mock a successful response for the AuthLogin call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
	}
	expectedRequests := []expectedRequest{{method: "POST", url: "/api/v2/auth/login"}}

	mockTransport := &mockRoundTripper{
		endpointResponses: endpointResponses,
		expectedRequests:  expectedRequests,
		requestIndex:      0,
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client, err := NewClient("testuser", "testpass", "localhost", "8080", httpClient)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.username != "testuser" || client.password != "testpass" {
		t.Errorf("Username or password not set correctly")
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestAuthLogin(t *testing.T) {
	// Mock a successful response for the AuthLogin call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "POST", url: "/api/v2/auth/login"},
	}

	mockTransport := &mockRoundTripper{
		endpointResponses: endpointResponses,
		expectedRequests:  expectedRequests,
		requestIndex:      0,
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client, err := NewClient("testuser", "testpass", "localhost", "8080", httpClient)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.AuthLogin()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestAuthLogin_Failure(t *testing.T) {
	// Mock a failure response for the AuthLogin call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusUnauthorized, responseBody: "Unauthorized"},
	}
	expectedRequests := []expectedRequest{{method: "POST", url: "/api/v2/auth/login"}}

	mockTransport := &mockRoundTripper{
		endpointResponses: endpointResponses,
		expectedRequests:  expectedRequests,
		requestIndex:      0,
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client, err := NewClient("testuser", "testpass", "localhost", "8080", httpClient)
	if err == nil || client != nil {
		t.Fatalf("Expected error during NewClient creation, got client: %v, error: %v", client, err)
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}
