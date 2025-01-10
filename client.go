package qbittorrent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type InfoHash string

// Client is used to interact with the qBittorrent API
type Client struct {
	username string
	password string
	client   *http.Client
	baseURL  string
	sid      string // store the SID cookie
	mu       sync.RWMutex
}

// TorrentInfo represents the structured information of a torrent from the qBittorrent API
type TorrentInfo struct {
	AddedOn            int64    `json:"added_on"`
	AmountLeft         int64    `json:"amount_left"`
	AutoTMM            bool     `json:"auto_tmm"`
	Availability       float64  `json:"availability"`
	Category           string   `json:"category"`
	Completed          int64    `json:"completed"`
	CompletionOn       int64    `json:"completion_on"`
	ContentPath        string   `json:"content_path"`
	DLLimit            int64    `json:"dl_limit"`
	DLSpeed            int64    `json:"dlspeed"`
	Downloaded         int64    `json:"downloaded"`
	DownloadedSession  int64    `json:"downloaded_session"`
	ETA                int64    `json:"eta"`
	FirstLastPiecePrio bool     `json:"f_l_piece_prio"`
	ForceStart         bool     `json:"force_start"`
	Hash               InfoHash `json:"hash"`
	IsPrivate          bool     `json:"isPrivate"`
	LastActivity       int64    `json:"last_activity"`
	MagnetURI          string   `json:"magnet_uri"`
	MaxRatio           float64  `json:"max_ratio"`
	MaxSeedingTime     int64    `json:"max_seeding_time"`
	Name               string   `json:"name"`
	NumComplete        int64    `json:"num_complete"`
	NumIncomplete      int64    `json:"num_incomplete"`
	NumLeechs          int64    `json:"num_leechs"`
	NumSeeds           int64    `json:"num_seeds"`
	Priority           int64    `json:"priority"`
	Progress           float64  `json:"progress"`
	Ratio              float64  `json:"ratio"`
	RatioLimit         float64  `json:"ratio_limit"`
	SavePath           string   `json:"save_path"`
	SeedingTime        int64    `json:"seeding_time"`
	SeedingTimeLimit   int64    `json:"seeding_time_limit"`
	SeenComplete       int64    `json:"seen_complete"`
	SequentialDownload bool     `json:"seq_dl"`
	Size               int64    `json:"size"`
	State              string   `json:"state"`
	SuperSeeding       bool     `json:"super_seeding"`
	Tags               []string `json:"-"`
	TimeActive         int64    `json:"time_active"`
	TotalSize          int64    `json:"total_size"`
	Tracker            string   `json:"tracker"`
	UpLimit            int64    `json:"up_limit"`
	Uploaded           int64    `json:"uploaded"`
	UploadedSession    int64    `json:"uploaded_session"`
	UpSpeed            int64    `json:"upspeed"`
}

// UnmarshalJSON custom unmarshaller for TorrentInfo to handle Tags
func (t *TorrentInfo) UnmarshalJSON(data []byte) error {
	type Alias TorrentInfo
	aux := &struct {
		RawTags string `json:"tags"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.RawTags == "" {
		t.Tags = []string{}
	} else {
		t.Tags = strings.Split(aux.RawTags, ",")
	}
	return nil
}

// TrackerInfo represents a tracker info for a torrent
type TrackerInfo struct {
	URL      string `json:"url"`
	Status   int    `json:"status"`
	Tier     int    `json:"tier"`
	NumPeers int    `json:"num_peers"`
	Msg      string `json:"msg"`
}

type Category map[string]interface{} // no idea what this should be, category=CategoryName&savePath=/path/to/dir

// fields might be missing, in which case we need to switch to pointers and allow "omitempty"
// https://github.com/qbittorrent/qBittorrent/blob/master/src/base/json_api.cpp#L101
// MainData is the data returned by the /api/v2/sync/maindata endpoint
type MainData struct {
	Categories        map[string]Category    `json:"categories"`
	CategoriesRemoved []Category             `json:"categories_removed"`
	FullUpdate        bool                   `json:"full_update"`
	Rid               int                    `json:"rid"`
	ServerState       ServerState            `json:"server_state"`
	Tags              []string               `json:"tags"`
	TagsRemoved       []string               `json:"tags_removed"`
	Torrents          map[string]TorrentInfo `json:"torrents"`
	TorrentsRemoved   []string               `json:"torrents_removed"`
	Trackers          map[string][]InfoHash  `json:"trackers"` // maps trackers to infohashes
}

type ServerState struct {
	AllTimeDL            int64  `json:"alltime_dl"`
	AllTimeUL            int64  `json:"alltime_ul"`
	AverageTimeQueue     int    `json:"average_time_queue"`
	ConnectionStatus     string `json:"connection_status"`
	DHTNodes             int    `json:"dht_nodes"`
	DLInfoData           int64  `json:"dl_info_data"`
	DLInfoSpeed          int    `json:"dl_info_speed"`
	DLRateLimit          int    `json:"dl_rate_limit"`
	FreeSpaceOnDisk      int64  `json:"free_space_on_disk"`
	GlobalRatio          string `json:"global_ratio"`
	QueuedIOJobs         int    `json:"queued_io_jobs"`
	Queueing             bool   `json:"queueing"`
	ReadCacheHits        string `json:"read_cache_hits"`
	ReadCacheOverload    string `json:"read_cache_overload"`
	RefreshInterval      int    `json:"refresh_interval"`
	TotalBuffersSize     int64  `json:"total_buffers_size"`
	TotalPeerConnections int    `json:"total_peer_connections"`
	TotalQueuedSize      int64  `json:"total_queued_size"`
	TotalWastedSession   int64  `json:"total_wasted_session"`
	UpInfoData           int64  `json:"up_info_data"`
	UpInfoSpeed          int    `json:"up_info_speed"`
	UpRateLimit          int    `json:"up_rate_limit"`
	UseAltSpeedLimits    bool   `json:"use_alt_speed_limits"`
	UseSubcategories     bool   `json:"use_subcategories"`
	WriteCacheOverload   string `json:"write_cache_overload"`
}

type TorrentPeer struct {
	Client       string  `json:"client"`
	Connection   string  `json:"connection"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	DLSpeed      int64   `json:"dl_speed"`
	Downloaded   int64   `json:"downloaded"`
	Files        string  `json:"files"`
	Flags        string  `json:"flags"`
	FlagsDesc    string  `json:"flags_desc"`
	IP           string  `json:"ip"`
	PeerIDClient string  `json:"peer_id_client"`
	Port         int     `json:"port"`
	Progress     float64 `json:"progress"`
	Relevance    float64 `json:"relevance"`
	Uploaded     int64   `json:"uploaded"`
	UPSpeed      int64   `json:"up_speed"`
}

type TorrentPeers struct {
	FullUpdate bool                   `json:"full_update"`
	Peers      map[string]TorrentPeer `json:"peers"`
	// PeersRemoved map[string][]string    `json:"peers_removed"`
	Rid       int  `json:"rid"`
	ShowFlags bool `json:"show_flags"`
}

// NewClient initializes a new qBittorrent client.
// If httpClient is nil, http.DefaultClient is used.
func NewClient(username, password, addr, port string, httpClient ...*http.Client) (*Client, error) {
	// Use the provided http.Client if given, otherwise use http.DefaultClient
	client := http.DefaultClient
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}

	// Create and return the Client instance
	qbClient := &Client{
		username: username,
		password: password,
		client:   client,
		baseURL:  fmt.Sprintf("http://%s:%s", addr, port),
	}

	// Authenticate if username and password are provided
	if username != "" && password != "" {
		if err := qbClient.AuthLogin(); err != nil {
			return nil, fmt.Errorf("AuthLogin error: %v", err)
		}
	}

	return qbClient, nil
}

// AuthLogin logs in to the qBittorrent Web API
func (c *Client) AuthLogin() error {
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("password", c.password)

	resp, err := c.doPostResponse("/api/v2/auth/login", strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return fmt.Errorf("AuthLogin error: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AuthLogin error (%d): %s", resp.StatusCode, string(respBody))
	}
	defer resp.Body.Close()

	// Extract the SID cookie from the response
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			c.mu.Lock()
			c.sid = cookie.Value
			c.mu.Unlock()
			break
		}
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

// These are not all the options, just the ones i need
// documentation at: https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#add-new-torrent
type TorrentsAddOptions struct {
	SavePath    *string
	Category    *string
	Tags        *[]string
	StartPaused *bool
	AutoTMM     *bool
}

type TorrentAddOption func(*TorrentsAddOptions)

func (c *Client) TorrentsAddWithOptions(torrentFile string, fileData []byte, opts ...TorrentAddOption) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Default this to true
	_ = writer.WriteField("skip_checking", "true") // Avoid recheck

	part, err := writer.CreateFormFile("torrents", torrentFile)
	if err != nil {
		return fmt.Errorf("CreateFormFile error: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(fileData)); err != nil {
		return fmt.Errorf("io.Copy error: %v", err)
	}

	options := &TorrentsAddOptions{}

	for _, opt := range opts {
		opt(options)
	}

	if options.SavePath != nil {
		_ = writer.WriteField("savepath", *options.SavePath)
	}

	if options.Category != nil {
		_ = writer.WriteField("category", *options.Category)
	}

	if options.Tags != nil {
		_ = writer.WriteField("tags", strings.Join(*options.Tags, ","))
	}

	if options.StartPaused != nil {
		_ = writer.WriteField("paused", strconv.FormatBool(*options.StartPaused))
	}

	if options.AutoTMM != nil {
		_ = writer.WriteField("autoTMM", strconv.FormatBool(*options.AutoTMM))
	}

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

// TorrentsInfoParams holds the optional parameters for the TorrentsInfo method
type TorrentsInfoParams struct {
	Filter   string
	Category string
	Tag      string
	Sort     string
	Reverse  bool
	Limit    int
	Offset   int
	Hashes   []string
}

// TorrentsInfo retrieves a list of all torrents from the qBittorrent server
func (c *Client) TorrentsInfo(params ...*TorrentsInfoParams) ([]TorrentInfo, error) {
	var query url.Values
	if len(params) > 0 && params[0] != nil {
		query = url.Values{}
		if params[0].Filter != "" {
			query.Set("filter", params[0].Filter)
		}
		if params[0].Category != "" {
			query.Set("category", params[0].Category)
		}
		if params[0].Tag != "" {
			query.Set("tag", params[0].Tag)
		}
		if params[0].Sort != "" {
			query.Set("sort", params[0].Sort)
		}
		if params[0].Reverse {
			query.Set("reverse", "true")
		}
		if params[0].Limit > 0 {
			query.Set("limit", strconv.Itoa(params[0].Limit))
		}
		if params[0].Offset != 0 {
			query.Set("offset", strconv.Itoa(params[0].Offset))
		}
		if len(params[0].Hashes) > 0 {
			query.Set("hashes", strings.Join(params[0].Hashes, "|"))
		}
	}

	respData, err := c.doGet("/api/v2/torrents/info", query)
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

// TorrentsAddTags adds tags to the specified torrents
func (c *Client) TorrentsAddTags(hashes, tags string) error {
	data := url.Values{}
	data.Set("hashes", hashes)
	data.Set("tags", tags)

	_, err := c.doPostValues("/api/v2/torrents/addTags", data)
	if err != nil {
		return fmt.Errorf("AddTags error: %v", err)
	}
	return nil
}

// TorrentsRemoveTags removes tags from the specified torrents
func (c *Client) TorrentsRemoveTags(hashes, tags string) error {
	data := url.Values{}
	data.Set("hashes", hashes)
	data.Set("tags", tags)

	_, err := c.doPostValues("/api/v2/torrents/removeTags", data)
	if err != nil {
		return fmt.Errorf("RemoveTags error: %v", err)
	}
	return nil
}

// TorrentsGetTags retrieves the tags for the given torrent hashes
func (c *Client) TorrentsGetTags(hashes string) ([]string, error) {
	params := &TorrentsInfoParams{
		Hashes: []string{hashes},
	}

	torrents, err := c.TorrentsInfo(params)
	if err != nil {
		return nil, fmt.Errorf("TorrentsGetTags error: %v", err)
	}

	tagSet := make(map[string]struct{})
	for _, torrent := range torrents {
		for _, tag := range torrent.Tags {
			tagSet[tag] = struct{}{}
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}

// TorrentsGetAllTags retrieves all tags from qBittorrent
func (c *Client) TorrentsGetAllTags() ([]string, error) {
	respData, err := c.doGet("/api/v2/torrents/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("GetAllTags error: %v", err)
	}

	var tags []string
	if err := json.Unmarshal(respData, &tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags response: %v", err)
	}

	return tags, nil
}

// TorrentsCreateTags creates new tags in qBittorrent
func (c *Client) TorrentsCreateTags(tags string) error {
	data := url.Values{}
	data.Set("tags", tags)

	_, err := c.doPostValues("/api/v2/torrents/createTags", data)
	if err != nil {
		return fmt.Errorf("CreateTags error: %v", err)
	}
	return nil
}

// TorrentsDeleteTags deletes tags from qBittorrent
func (c *Client) TorrentsDeleteTags(tags string) error {
	data := url.Values{}
	data.Set("tags", tags)

	_, err := c.doPostValues("/api/v2/torrents/deleteTags", data)
	if err != nil {
		return fmt.Errorf("DeleteTags error: %v", err)
	}
	return nil
}

// doPostResponse POSTs to qBittorrent and returns the HTTP response
func (c *Client) doPostResponse(endpoint string, body io.Reader, contentType string) (*http.Response, error) {
	return c.doRequest("POST", endpoint, body, contentType)
}

// doPost makes POSTs to qBittorrent and returns the response body
func (c *Client) doPost(endpoint string, body io.Reader, contentType string) ([]byte, error) {
	resp, err := c.doPostResponse(endpoint, body, contentType)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST error (%d): %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// doPostValues POSTs to qBittorrent with url.Values and returns the response body
func (c *Client) doPostValues(endpoint string, data url.Values) ([]byte, error) {
	return c.doPost(endpoint, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
}

// doGet is a helper method for making GET requests to the qBittorrent API with query parameters
func (c *Client) doGet(endpoint string, query url.Values) ([]byte, error) {
	resp, err := c.doRequest("GET", endpoint, nil, "", withQuery(query))
	if err != nil {
		return nil, err
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

// doRequest is a helper function to handle HTTP requests with optional query parameters
func (c *Client) doRequest(method, endpoint string, body io.Reader, contentType string, opts ...func(*http.Request) error) (*http.Response, error) {
	apiURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %v", err)
	}

	apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + endpoint

	// Store body in buffer if it's not nil so we can retry the request
	var bodyBuffer []byte
	if body != nil {
		bodyBuffer, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
	}

	makeRequest := func() (*http.Request, error) {
		var bodyReader io.Reader
		if bodyBuffer != nil {
			bodyReader = bytes.NewReader(bodyBuffer)
		}
		req, err := http.NewRequest(method, apiURL.String(), bodyReader)
		if err != nil {
			return nil, fmt.Errorf("NewRequest error: %v", err)
		}

		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		c.mu.RLock()
		if c.sid != "" {
			req.AddCookie(&http.Cookie{Name: "SID", Value: c.sid})
		}
		c.mu.RUnlock()

		// Apply any optional request modifiers
		for _, opt := range opts {
			if err := opt(req); err != nil {
				return nil, err
			}
		}
		return req, nil
	}

	// Make initial request
	req, err := makeRequest()
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// If we get a 403 Forbidden, try to re-authenticate once and retry the request
	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close() // Close the first response

		if err := c.AuthLogin(); err != nil {
			return nil, fmt.Errorf("re-authentication failed: %v", err)
		}

		// Retry the original request with the new SID
		req, err := makeRequest()
		if err != nil {
			return nil, err
		}

		return c.client.Do(req)
	}

	return resp, nil
}

// withQuery returns a request modifier that adds query parameters
func withQuery(query url.Values) func(*http.Request) error {
	return func(req *http.Request) error {
		req.URL.RawQuery = query.Encode()
		return nil
	}
}

func (c *Client) SyncMainData(rid int) (*MainData, error) {
	params := url.Values{}
	params.Set("rid", strconv.Itoa(rid))

	resp, err := c.doGet("/api/v2/sync/maindata", params)
	if err != nil {
		return nil, err
	}

	var result MainData
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) SyncTorrentPeers(hash string, rid int) (*TorrentPeers, error) {
	params := url.Values{}
	params.Set("rid", strconv.Itoa(rid))
	params.Set("hash", hash)

	resp, err := c.doGet("/api/v2/sync/torrentPeers", params)
	if err != nil {
		return nil, err
	}

	var result TorrentPeers
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
