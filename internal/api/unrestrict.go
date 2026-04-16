package api

import (
	"fmt"
	"net/url"
)

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

// LinkError represents an error that occurred while unrestricting a specific link.
type LinkError struct {
	Link string
	Err  error
}

func (e *LinkError) Error() string {
	return fmt.Sprintf("unrestrict %s: %s", e.Link, e.Err)
}

func (e *LinkError) Unwrap() error {
	return e.Err
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

// UnrestrictLinks unrestricts multiple links, collecting results and errors separately.
func UnrestrictLinks(client *Client, links []string) ([]*UnrestrictedLink, []*LinkError) {
	var results []*UnrestrictedLink
	var errs []*LinkError

	for _, link := range links {
		result, err := UnrestrictLink(client, link)
		if err != nil {
			errs = append(errs, &LinkError{Link: link, Err: err})
		} else {
			results = append(results, result)
		}
	}

	return results, errs
}
