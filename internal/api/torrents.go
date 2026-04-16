package api

import (
	"fmt"
	"net/url"
	"strings"
)

// TorrentInfo represents a torrent entry from the Real-Debrid API.
type TorrentInfo struct {
	ID       string        `json:"id"`
	Filename string        `json:"filename"`
	Hash     string        `json:"hash"`
	Bytes    int64         `json:"bytes"`
	Host     string        `json:"host"`
	Status   string        `json:"status"`
	Progress int           `json:"progress"`
	Links    []string      `json:"links"`
	Files    []TorrentFile `json:"files"`
}

// TorrentFile represents a file within a torrent.
type TorrentFile struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Bytes    int64  `json:"bytes"`
	Selected int    `json:"selected"`
}

// AddTorrentResponse represents the response from adding a magnet link.
type AddTorrentResponse struct {
	ID  string `json:"id"`
	URI string `json:"uri"`
}

// ListTorrents returns the user's torrent list.
func ListTorrents(client *Client) ([]TorrentInfo, error) {
	var result []TorrentInfo
	if err := client.Get("/torrents", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTorrentInfo returns detailed information about a specific torrent.
func GetTorrentInfo(client *Client, id string) (*TorrentInfo, error) {
	var result TorrentInfo
	if err := client.Get("/torrents/info/"+id, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddMagnet adds a magnet link to the user's torrents.
func AddMagnet(client *Client, magnet string) (*AddTorrentResponse, error) {
	form := url.Values{"magnet": {magnet}}
	var result AddTorrentResponse
	if err := client.Post("/torrents/addMagnet", form, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SelectTorrentFiles selects which files to download from a torrent.
// fileIDs is a comma-separated list of file IDs, or "all" to select all files.
func SelectTorrentFiles(client *Client, id string, fileIDs string) error {
	form := url.Values{"files": {fileIDs}}
	return client.Post("/torrents/selectFiles/"+id, form, nil)
}

// DeleteTorrent removes a torrent from the user's list.
func DeleteTorrent(client *Client, id string) error {
	return client.Delete("/torrents/delete/" + id)
}

// FormatTorrentStatus maps API status strings to human-readable labels.
func FormatTorrentStatus(status string, progress int) string {
	switch strings.ToLower(status) {
	case "magnet_error":
		return "Magnet Error"
	case "magnet_conversion":
		return "Converting Magnet..."
	case "waiting_files_selection":
		return "Waiting for File Selection"
	case "queued":
		return "Queued"
	case "downloading":
		return fmt.Sprintf("Downloading (%d%%)", progress)
	case "downloaded":
		return "Downloaded"
	case "error":
		return "Error"
	case "virus":
		return "Virus Detected"
	case "compressing":
		return fmt.Sprintf("Compressing (%d%%)", progress)
	case "uploading":
		return fmt.Sprintf("Uploading (%d%%)", progress)
	case "dead":
		return "Dead"
	default:
		return status
	}
}
