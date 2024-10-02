
# qBittorrent Go Client Library

A Go client library for interacting with the [qBittorrent](https://www.qbittorrent.org/) Web API.

## Features

- **Authentication**: Log in to the qBittorrent Web API.
- **Torrent Management**:
  - Add new torrents.
  - Delete existing torrents.
  - Export torrent files.
  - Retrieve torrent information.
  - Manage torrent force-start settings.
- **Tracker Information**: Fetch tracker details for specific torrents.

## Installation

To install the package, run:

```bash
go get github.com/cehbz/qbittorrent
```

## Usage

### Importing the Package

```go
import (
    "github.com/cehbz/qbittorrent"
)
```

### Initializing the Client

```go
client, err := qbittorrent.NewClient("username", "password", "localhost", "8080")
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

- `username`: Your qBittorrent Web UI username. Empty if none.
- `password`: Your qBittorrent Web UI password. Empty if none.
- `addr`: The address where qBittorrent is running (e.g., `"127.0.0.1"`).
- `port`: The port number of the qBittorrent Web UI (e.g., `"8080"`).

### Adding a Torrent

```go
torrentData, err := os.ReadFile("path/to/your.torrent")
if err != nil {
    log.Fatalf("Failed to read torrent file: %v", err)
}

err = client.TorrentsAdd("your.torrent", torrentData)
if err != nil {
    log.Fatalf("Failed to add torrent: %v", err)
}
```

### Deleting a Torrent

```go
err := client.TorrentsDelete("torrent-hash")
if err != nil {
    log.Fatalf("Failed to delete torrent: %v", err)
}
```

### Exporting a Torrent File

```go
data, err := client.TorrentsExport("torrent-hash")
if err != nil {
    log.Fatalf("Failed to export torrent: %v", err)
}

err = os.WriteFile("exported.torrent", data, 0644)
if err != nil {
    log.Fatalf("Failed to write exported torrent file: %v", err)
}
```

### Retrieving Torrent Information

```go
torrents, err := client.TorrentsInfo()
if err != nil {
    log.Fatalf("Failed to retrieve torrents info: %v", err)
}

for _, torrent := range torrents {
    fmt.Printf("Name: %s, Progress: %.2f%%
", torrent.Name, torrent.Progress*100)
}
```

### Fetching Tracker Information

```go
trackers, err := client.TorrentsTrackers("torrent-hash")
if err != nil {
    log.Fatalf("Failed to get trackers: %v", err)
}

for _, tracker := range trackers {
    fmt.Printf("URL: %s, Status: %d
", tracker.URL, tracker.Status)
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contribution

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Acknowledgments

- [qBittorrent Web API Documentation](https://github.com/qbittorrent/qBittorrent/wiki/Web-API-Documentation)
