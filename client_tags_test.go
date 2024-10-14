package qbittorrent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTorrentInfo_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected []string
	}{
		{
			name:     "Empty tags",
			jsonData: `{"tags": ""}`,
			expected: []string{},
		},
		{
			name:     "One tag",
			jsonData: `{"tags": "tag1"}`,
			expected: []string{"tag1"},
		},
		{
			name:     "Multiple tags",
			jsonData: `{"tags": "tag1,tag2,tag3"}`,
			expected: []string{"tag1", "tag2", "tag3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var torrentInfo TorrentInfo
			err := json.Unmarshal([]byte(tt.jsonData), &torrentInfo)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(torrentInfo.Tags) != len(tt.expected) {
				t.Fatalf("expected %d tags, got %d", len(tt.expected), len(torrentInfo.Tags))
			}

			for i, tag := range tt.expected {
				if torrentInfo.Tags[i] != tag {
					t.Errorf("expected tag %v, got %v", tag, torrentInfo.Tags[i])
				}
			}
		})
	}
}

func TestClient_TorrentsGetTags(t *testing.T) {
	// Mock server response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"tags": "tag1,tag2"},{"tags": "tag2,tag3"}]`))
	}))
	defer mockServer.Close()

	client := &Client{
		baseURL: mockServer.URL,
		client:  mockServer.Client(),
	}

	tags, err := client.TorrentsGetTags("somehash1|somehash2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedTags := []string{"tag1", "tag2", "tag3"}
	if len(tags) != len(expectedTags) {
		t.Fatalf("expected %d tags, got %d", len(expectedTags), len(tags))
	}

	tagSet := make(map[string]struct{})
	for _, tag := range tags {
		tagSet[tag] = struct{}{}
	}

	for _, expectedTag := range expectedTags {
		if _, exists := tagSet[expectedTag]; !exists {
			t.Errorf("expected tag %v to be present", expectedTag)
		}
	}
}
