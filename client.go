package qbittorrent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// Client is used to interact with the qBittorrent API
type Client struct {
	username string
	password string
	addr     string
	port     string
	client   *http.Client
}

// TorrentInfo represents the structured information of a torrent from the qBittorrent API
type TorrentInfo struct {
	AddedOn            int64   `json:"added_on"`
	Name               string  `json:"name"`
	State              string  `json:"state"`
	Hash               string  `json:"hash"`
	LastActivity       int64   `json:"last_activity"`
	Progress           float64 `json:"progress"`
	Downloaded         int64   `json:"downloaded"`
	Uploaded           int64   `json:"uploaded"`
	Size               int64   `json:"size"`
	Category           string  `json:"category"`
	SavePath           string  `json:"save_path"`
	CompletionOn       int64   `json:"completion_on"`
	DLSpeed            int64   `json:"dlspeed"`
	UpSpeed            int64   `json:"upspeed"`
	AmountLeft         int64   `json:"amount_left"`
	AutoTMM            bool    `json:"auto_tmm"`
	Availability       float64 `json:"availability"`
	Completed          int64   `json:"completed"`
	ContentPath        string  `json:"content_path"`
	DLLimit            int64   `json:"dl_limit"`
	DownloadedSession  int64   `json:"downloaded_session"`
	ETA                int64   `json:"eta"`
	FirstLastPiecePrio bool    `json:"f_l_piece_prio"`
	ForceStart         bool    `json:"force_start"`
	IsPrivate          bool    `json:"isPrivate"`
	MagnetURI          string  `json:"magnet_uri"`
	MaxRatio           float64 `json:"max_ratio"`
	MaxSeedingTime     int64   `json:"max_seeding_time"`
	NumComplete        int64   `json:"num_complete"`
	NumIncomplete      int64   `json:"num_incomplete"`
	NumLeechs          int64   `json:"num_leechs"`
	NumSeeds           int64   `json:"num_seeds"`
	Priority           int64   `json:"priority"`
	Ratio              float64 `json:"ratio"`
	RatioLimit         float64 `json:"ratio_limit"`
	SeedingTime        int64   `json:"seeding_time"`
	SeedingTimeLimit   int64   `json:"seeding_time_limit"`
	SeenComplete       int64   `json:"seen_complete"`
	SequentialDownload bool    `json:"seq_dl"`
	SuperSeeding       bool    `json:"super_seeding"`
	Tags               string  `json:"tags"`
	TimeActive         int64   `json:"time_active"`
	TotalSize          int64   `json:"total_size"`
	Tracker            string  `json:"tracker"`
	UpLimit            int64   `json:"up_limit"`
	UploadedSession    int64   `json:"uploaded_session"`
}

// TrackerInfo represents a tracker info for a torrent
type TrackerInfo struct {
	URL      string `json:"url"`
	Status   int    `json:"status"`
	Tier     int    `json:"tier"`
	NumPeers int    `json:"num_peers"`
	Msg      string `json:"msg"`
}

// NewClient initializes a new qBittorrent client
func NewClient(username, password, addr, port string) (*Client, error) {
	client := &Client{
		username: username,
		password: password,
		addr:     addr,
		port:     port,
		client:   &http.Client{}, // Using default http.Client
	}

	// Authenticate if username and password are provided
	if username != "" && password != "" {
		if err := client.AuthLogin(); err != nil {
			return nil, fmt.Errorf("AuthLogin error: %v", err)
		}
	}
	return client, nil
}

// AuthLogin logs in to the qBittorrent Web API
func (c *Client) AuthLogin() error {
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("password", c.password)

	_, err := c.doPostValues("/api/v2/auth/login", data)
	if err != nil {
		return fmt.Errorf("AuthLogin error: %v", err)
	}
	return nil
}

// TorrentsExport retrieves the .torrent file for a given torrent hash
func (c *Client) TorrentsExport(hash string) ([]byte, error) {
	params := url.Values{}
	params.Set("hash", hash)

	// Use the GET request helper
	return c.doPostValues("/api/v2/torrents/export", params)
}

// TorrentsAdd adds a torrent to qBittorrent via Web API using multipart/form-data
func (c *Client) TorrentsAdd(torrentFile string, fileData []byte) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("torrents", torrentFile)
	if err != nil {
		return fmt.Errorf("CreateFormFile error: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(fileData)); err != nil {
		return fmt.Errorf("io.Copy error: %v", err)
	}

	_ = writer.WriteField("skip_checking", "true") // Avoid recheck
	_ = writer.WriteField("paused", "false")
	_ = writer.WriteField("autoTMM", "false")
	writer.Close()

	_, err = c.doPost("/api/v2/torrents/add", &body, writer.FormDataContentType())
	if err != nil {
		return fmt.Errorf("TorrentsAdd error: %v", err)
	}
	return nil
}

// TorrentsDelete deletes a torrent from qBittorrent by its hash
func (c *Client) TorrentsDelete(infohash string) error {
	data := url.Values{}
	data.Set("hashes", infohash)
	data.Set("deleteFiles", "true")

	_, err := c.doPostValues("/api/v2/torrents/delete", data)
	if err != nil {
		return fmt.Errorf("TorrentsDelete error: %v", err)
	}
	return nil
}

// SetForceStart enables force start for the torrent
func (c *Client) SetForceStart(hash string, value bool) error {
	data := url.Values{}
	data.Set("hashes", hash)
	data.Set("value", fmt.Sprintf("%t", value))

	_, err := c.doPostValues("/api/v2/torrents/setForceStart", data)
	if err != nil {
		return fmt.Errorf("SetForceStart error: %v", err)
	}
	return nil
}

// TorrentsDownload retrieves the torrent file by its hash from the qBittorrent server
func (c *Client) TorrentsDownload(infohash string) ([]byte, error) {
	return c.doGet("/api/v2/torrents/file", url.Values{"hashes": {infohash}})
}

// TorrentsInfo retrieves a list of all torrents from the qBittorrent server
func (c *Client) TorrentsInfo() ([]TorrentInfo, error) {
	respData, err := c.doGet("/api/v2/torrents/info", nil)
	if err != nil {
		return nil, err
	}

	var torrents []TorrentInfo
	if err := json.Unmarshal(respData, &torrents); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return torrents, nil
}

// TorrentsTrackers retrieves the tracker info for a given torrent hash
func (c *Client) TorrentsTrackers(hash string) ([]TrackerInfo, error) {
	params := url.Values{}
	params.Set("hash", hash)

	respData, err := c.doGet("/api/v2/torrents/trackers", params)
	if err != nil {
		return nil, fmt.Errorf("TorrentsTrackers error: %v", err)
	}

	var trackers []TrackerInfo
	if err := json.Unmarshal(respData, &trackers); err != nil {
		return nil, fmt.Errorf("failed to decode trackers response: %v", err)
	}

	return trackers, nil
}

// doPostValues is a helper method for making POST requests to the qBittorrent API with url.Values
func (c *Client) doPostValues(endpoint string, data url.Values) ([]byte, error) {
	return c.doPost(endpoint, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
}

// doPost is a helper method for making POST requests to the qBittorrent API
func (c *Client) doPost(endpoint string, body io.Reader, contentType string) ([]byte, error) {
	apiURL := fmt.Sprintf("http://%s:%s%s", c.addr, c.port, endpoint)

	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("NewRequest error: %v", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll error: %v", err)
	}
	return responseData, nil
}

// doGet is a helper method for making GET requests to the qBittorrent API with query parameters
func (c *Client) doGet(endpoint string, query url.Values) ([]byte, error) {
	apiURL := fmt.Sprintf("http://%s:%s%s", c.addr, c.port, endpoint)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("NewRequest error: %v", err)
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll error: %v", err)
	}
	return responseData, nil
}
