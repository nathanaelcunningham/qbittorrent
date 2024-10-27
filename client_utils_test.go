package qbittorrent

import (
	"bytes"
	"io"
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

func TestDoRequest(t *testing.T) {
	tests := []struct {
		name              string
		endpointResponses map[string]mockResponse
		expectedRequests  []expectedRequest
		method            string
		endpoint          string
		body              io.Reader
		contentType       string
		query             url.Values
		wantErr           bool
		wantResponse      string
	}{
		{
			name: "successful request with SID",
			endpointResponses: map[string]mockResponse{
				"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
				"/api/test":          {statusCode: http.StatusOK, responseBody: "Success"},
			},
			expectedRequests: []expectedRequest{
				{method: "POST", url: "/api/v2/auth/login"},
				{method: "GET", url: "/api/test"},
			},
			method:       "GET",
			endpoint:     "/api/test",
			wantResponse: "Success",
		},
		{
			name: "request with query parameters",
			endpointResponses: map[string]mockResponse{
				"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
				"/api/test":          {statusCode: http.StatusOK, responseBody: "Success"},
			},
			expectedRequests: []expectedRequest{
				{method: "POST", url: "/api/v2/auth/login"},
				{method: "GET", url: "/api/test", query: url.Values{"key": []string{"value"}}},
			},
			method:       "GET",
			endpoint:     "/api/test",
			query:        url.Values{"key": []string{"value"}},
			wantResponse: "Success",
		},
		{
			name: "403 with successful reauth",
			endpointResponses: map[string]mockResponse{
				"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
				"/api/test": {
					statusCode:   http.StatusForbidden,
					responseBody: "Forbidden",
					then: &mockResponse{
						statusCode:   http.StatusOK,
						responseBody: "Success after reauth",
					},
				},
			},
			expectedRequests: []expectedRequest{
				{method: "POST", url: "/api/v2/auth/login"},
				{method: "GET", url: "/api/test"},
				{method: "POST", url: "/api/v2/auth/login"},
				{method: "GET", url: "/api/test"},
			},
			method:       "GET",
			endpoint:     "/api/test",
			wantResponse: "Success after reauth",
		},
		{
			name: "403 with failed reauth",
			endpointResponses: map[string]mockResponse{
				"/api/v2/auth/login": {
					statusCode:   http.StatusOK,
					responseBody: "Ok.",
					then: &mockResponse{
						statusCode:   http.StatusUnauthorized,
						responseBody: "Auth failed",
					},
				},
				"/api/test": {statusCode: http.StatusForbidden, responseBody: "Forbidden"},
			},
			expectedRequests: []expectedRequest{
				{method: "POST", url: "/api/v2/auth/login"},
				{method: "GET", url: "/api/test"},
				{method: "POST", url: "/api/v2/auth/login"},
			},
			method:   "GET",
			endpoint: "/api/test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockTransport, err := newMockClient(tt.endpointResponses, tt.expectedRequests)
			if err != nil {
				t.Fatalf("Failed to create mock client: %v", err)
			}

			var opts []func(*http.Request) error
			if tt.query != nil {
				opts = append(opts, withQuery(tt.query))
			}

			resp, err := client.doRequest(tt.method, tt.endpoint, tt.body, tt.contentType, opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("doRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				if string(body) != tt.wantResponse {
					t.Errorf("doRequest() response = %v, want %v", string(body), tt.wantResponse)
				}
			}

			if mockTransport.requestIndex != len(mockTransport.expectedRequests) {
				t.Errorf("Not all expected requests were made. Made %d, expected %d",
					mockTransport.requestIndex, len(mockTransport.expectedRequests))
			}
		})
	}
}
