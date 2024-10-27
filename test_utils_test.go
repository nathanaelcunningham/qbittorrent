package qbittorrent

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// mockRoundTripper is used to mock http.Client responses and supports multiple endpoints
type mockRoundTripper struct {
	responses        map[string]mockResponse
	expectedRequests []expectedRequest
	requestIndex     int
	t                *testing.T
}

// mockResponse represents a mock HTTP response for a given endpoint
type mockResponse struct {
	statusCode   int
	responseBody string
	then         *mockResponse // Next response in sequence
}

// expectedRequest represents an expected HTTP request
type expectedRequest struct {
	method string
	url    string
	params url.Values // For POST form values
	query  url.Values // For GET query parameters
}

// RoundTrip implements the RoundTripper interface
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.requestIndex >= len(m.expectedRequests) {
		m.t.Errorf("Unexpected request: %s %s", req.Method, req.URL.Path)
		return nil, fmt.Errorf("unexpected request")
	}

	expected := m.expectedRequests[m.requestIndex]
	m.requestIndex++

	// Check method and path
	if req.Method != expected.method {
		m.t.Errorf("Expected method %s, got %s", expected.method, req.Method)
	}
	if req.URL.Path != expected.url {
		m.t.Errorf("Expected URL %s, got %s", expected.url, req.URL.Path)
	}

	// Check query parameters if specified
	if expected.query != nil {
		if !reflect.DeepEqual(req.URL.Query(), expected.query) {
			m.t.Errorf("Expected query params %v, got %v", expected.query, req.URL.Query())
		}
	}

	// Check POST form values if specified
	if expected.params != nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(req.PostForm, expected.params) {
			m.t.Errorf("Expected form values %v, got %v", expected.params, req.PostForm)
		}
	}

	resp := m.responses[req.URL.Path]
	// If there's a sequential response, update it for next time
	if resp.then != nil {
		m.responses[req.URL.Path] = *resp.then
	}

	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(strings.NewReader(resp.responseBody)),
		Header:     make(http.Header),
	}, nil
}

// helper function to create a mock client with predefined endpoint responses and expected requests
func newMockClient(responses map[string]mockResponse, expectedRequests []expectedRequest) (*Client, *mockRoundTripper, error) {
	transport := &mockRoundTripper{
		responses:        responses,
		expectedRequests: expectedRequests,
		t:                &testing.T{},
	}

	httpClient := &http.Client{Transport: transport}
	client, err := NewClient("user", "pass", "localhost", "8080", httpClient)
	return client, transport, err
}
