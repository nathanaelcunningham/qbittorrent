package qbittorrent

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

// mockRoundTripper is used to mock http.Client responses and supports multiple endpoints
type mockRoundTripper struct {
	endpointResponses map[string]mockResponse
	expectedRequests  []expectedRequest
	requestIndex      int
}

// mockResponse represents a mock HTTP response for a given endpoint
type mockResponse struct {
	statusCode   int
	responseBody string
	err          error
}

// expectedRequest represents an expected HTTP request
type expectedRequest struct {
	method string
	url    string
	params url.Values
}

// RoundTrip implements the RoundTripper interface
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.requestIndex >= len(m.expectedRequests) {
		return nil, http.ErrHandlerTimeout
	}

	expected := m.expectedRequests[m.requestIndex]
	if req.Method != expected.method || req.URL.Path != expected.url {
		return nil, &url.Error{Op: "Get", URL: req.URL.String(), Err: http.ErrAbortHandler}
	}

	if expected.params != nil {
		err := req.ParseForm()
		if err != nil {
			return nil, err
		}
		for key, values := range expected.params {
			if req.Form.Get(key) != values[0] {
				return nil, &url.Error{Op: "Get", URL: req.URL.String(), Err: http.ErrAbortHandler}
			}
		}
	}

	m.requestIndex++
	response, ok := m.endpointResponses[req.URL.Path]
	if !ok {
		// Default to 404 Not Found if endpoint is not defined
		response = mockResponse{
			statusCode:   http.StatusNotFound,
			responseBody: "Not Found",
		}
	}

	resp := &http.Response{
		StatusCode: response.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(response.responseBody)),
		Header:     make(http.Header),
	}

	return resp, response.err
}

// helper function to create a mock client with predefined endpoint responses and expected requests
func newMockClient(endpointResponses map[string]mockResponse, expectedRequests []expectedRequest) (*Client, *mockRoundTripper, error) {
	mockTransport := &mockRoundTripper{
		endpointResponses: endpointResponses,
		expectedRequests:  expectedRequests,
		requestIndex:      0,
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	// Directly return the client and error from NewClient
	client, err := NewClient("testuser", "testpass", "localhost", "8080", httpClient)
	return client, mockTransport, err
}
