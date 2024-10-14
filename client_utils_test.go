package qbittorrent

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

func TestDoPostValues(t *testing.T) {
	// Mock successful AuthLogin and generic POST response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusOK, responseBody: "Ok."},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "POST", url: "/api/test", params: url.Values{"key": []string{"value"}}},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data := url.Values{}
	data.Set("key", "value")

	resp, err := client.doPostValues("/api/test", data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(resp) != "Ok." {
		t.Errorf("Expected response 'Ok.', got '%s'", string(resp))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestDoPost_Error(t *testing.T) {
	// Mock successful AuthLogin and an error POST response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusInternalServerError, responseBody: "Internal Server Error"},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "POST", url: "/api/test"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data := bytes.NewBufferString("test data")
	_, err = client.doPost("/api/test", data, "text/plain")
	if err == nil {
		t.Fatalf("Expected error, got none")
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestDoGet(t *testing.T) {
	// Mock successful AuthLogin and a successful GET response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusOK, responseBody: "Response data"},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/test"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resp, err := client.doGet("/api/test", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(resp) != "Response data" {
		t.Errorf("Expected 'Response data', got '%s'", string(resp))
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}

func TestDoGet_Error(t *testing.T) {
	// Mock successful AuthLogin and an error GET response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusNotFound, responseBody: "Not Found"},
	}
	expectedRequests := []expectedRequest{
		{method: "POST", url: "/api/v2/auth/login"},
		{method: "GET", url: "/api/test"},
	}

	client, mockTransport, err := newMockClient(endpointResponses, expectedRequests)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = client.doGet("/api/test", nil)
	if err == nil {
		t.Fatalf("Expected error, got none")
	}

	// Check the request made
	if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
		t.Errorf("Not all expected requests were made")
	}
}
