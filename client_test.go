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
}

// mockResponse represents a mock HTTP response for a given endpoint
type mockResponse struct {
	statusCode   int
	responseBody string
	err          error
}

// RoundTrip implements the RoundTripper interface
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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

// helper function to create a mock client with predefined endpoint responses
func newMockClient(endpointResponses map[string]mockResponse) (*Client, error) {
	mockTransport := &mockRoundTripper{
		endpointResponses: endpointResponses,
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	// Directly return the client and error from NewClient
	return NewClient("testuser", "testpass", "localhost", "8080", httpClient)
}

func TestNewClient(t *testing.T) {
	// Mock a successful response for the AuthLogin call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
	}

	// Test without authentication
	client, err := NewClient("", "", "localhost", "8080")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.username != "" || client.password != "" {
		t.Errorf("Expected empty username and password")
	}

	// Test with authentication using mock client
	client, err = newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.username != "testuser" || client.password != "testpass" {
		t.Errorf("Username or password not set correctly")
	}
}

func TestAuthLogin(t *testing.T) {
	// Mock a successful response for the AuthLogin call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
	}

	client, err := newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.AuthLogin()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAuthLogin_Failure(t *testing.T) {
	// Mock a failure response for the AuthLogin call
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusUnauthorized, responseBody: "Unauthorized"},
	}

	client, err := newMockClient(endpointResponses)
	if err == nil || client != nil {
		t.Fatalf("Expected error during NewClient creation, got client: %v, error: %v", client, err)
	}
}

func TestTorrentsExport(t *testing.T) {
	expectedData := "torrent file data"
	// Mock successful AuthLogin and TorrentsExport responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":      {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/export": {statusCode: http.StatusOK, responseBody: expectedData},
	}

	client, err := newMockClient(endpointResponses)
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
}

func TestTorrentsAdd(t *testing.T) {
	// Mock successful AuthLogin and TorrentsAdd responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":   {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/add": {statusCode: http.StatusOK, responseBody: "Ok."},
	}

	client, err := newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.TorrentsAdd("test.torrent", []byte("torrent data"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestTorrentsDelete(t *testing.T) {
	// Mock successful AuthLogin and TorrentsDelete responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":      {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/delete": {statusCode: http.StatusOK, responseBody: "Ok."},
	}

	client, err := newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.TorrentsDelete("testhash")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSetForceStart(t *testing.T) {
	// Mock successful AuthLogin and SetForceStart responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":             {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/setForceStart": {statusCode: http.StatusOK, responseBody: "Ok."},
	}

	client, err := newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = client.SetForceStart("testhash", true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
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
	client, err := newMockClient(endpointResponses)
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

	// Test with parameters
	params := &TorrentsInfoParams{
		Filter:   "downloading",
		Category: "sample category",
		Sort:     "ratio",
	}
	torrents, err = client.TorrentsInfo(params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(torrents) != 2 {
		t.Errorf("Expected 2 torrents, got %d", len(torrents))
	}
}

func TestTorrentsTrackers(t *testing.T) {
	responseBody := `[{"url":"tracker1","status":1},{"url":"tracker2","status":0}]`
	// Mock successful AuthLogin and TorrentsTrackers responses
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login":        {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/v2/torrents/trackers": {statusCode: http.StatusOK, responseBody: responseBody},
	}

	client, err := newMockClient(endpointResponses)
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
}

func TestDoPostValues(t *testing.T) {
	// Mock successful AuthLogin and generic POST response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusOK, responseBody: "Ok."},
	}

	client, err := newMockClient(endpointResponses)
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
}

func TestDoPost_Error(t *testing.T) {
	// Mock successful AuthLogin and an error POST response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusInternalServerError, responseBody: "Internal Server Error"},
	}

	client, err := newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data := bytes.NewBufferString("test data")
	_, err = client.doPost("/api/test", data, "text/plain")
	if err == nil {
		t.Fatalf("Expected error, got none")
	}
}

func TestDoGet(t *testing.T) {
	// Mock successful AuthLogin and a successful GET response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusOK, responseBody: "Response data"},
	}

	client, err := newMockClient(endpointResponses)
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
}

func TestDoGet_Error(t *testing.T) {
	// Mock successful AuthLogin and an error GET response
	endpointResponses := map[string]mockResponse{
		"/api/v2/auth/login": {statusCode: http.StatusOK, responseBody: "Ok."},
		"/api/test":          {statusCode: http.StatusNotFound, responseBody: "Not Found"},
	}

	client, err := newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, err = client.doGet("/api/test", nil)
	if err == nil {
		t.Fatalf("Expected error, got none")
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
	client, err := newMockClient(endpointResponses)
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
	client, err := newMockClient(endpointResponses)
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
	client, err := newMockClient(endpointResponses)
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
	client, err := newMockClient(endpointResponses)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test with multiple hashes
	params := &TorrentsInfoParams{
		Hashes: []string{"hash1", "hash2"},
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
}
