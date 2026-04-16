package api

import "net/url"

// UnrestrictedLink represents a direct download link returned by the Real-Debrid API.
type UnrestrictedLink struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Filesize int64  `json:"filesize"`
	Download string `json:"download"`
	MimeType string `json:"mimeType"`
	Host     string `json:"host"`
	Chunks   int    `json:"chunks"`
}

// UnrestrictLink converts a hoster URL to a direct download URL via the Real-Debrid API.
func UnrestrictLink(client *Client, link string) (*UnrestrictedLink, error) {
	form := url.Values{"link": {link}}
	var result UnrestrictedLink
	if err := client.Post("/unrestrict/link", form, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

