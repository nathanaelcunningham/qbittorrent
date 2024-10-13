package qbittorrent

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
