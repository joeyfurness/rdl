package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	OAuthBaseURL    = "https://api.real-debrid.com/oauth/v2"
	DefaultClientID = "X245A4XAIBGVM"
)

// TokenData holds OAuth token information for persistence.
type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	ObtainedAt   int64  `json:"obtained_at"`
}

// IsExpired returns true if the token has expired (with a 5-minute buffer).
func (t *TokenData) IsExpired() bool {
	expiry := t.ObtainedAt + int64(t.ExpiresIn) - 300
	return time.Now().Unix() > expiry
}

// DeviceCodeResponse holds the response from the device code request.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceCredentials holds the client credentials returned after device authorization.
type DeviceCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// SaveTokens marshals tokens to JSON and writes them with 0600 permissions.
func SaveTokens(path string, tokens *TokenData) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling tokens: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// LoadTokens reads and unmarshals tokens from a JSON file.
func LoadTokens(path string) (*TokenData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var tokens TokenData
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("unmarshaling tokens: %w", err)
	}

	return &tokens, nil
}

// DeleteTokens removes the token file.
func DeleteTokens(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing token file: %w", err)
	}
	return nil
}

// RequestDeviceCode initiates the OAuth device flow by requesting a device code.
func RequestDeviceCode(baseURL, clientID string) (*DeviceCodeResponse, error) {
	reqURL := fmt.Sprintf("%s/device/code?client_id=%s&new_credentials=yes",
		baseURL, url.QueryEscape(clientID))

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var dcr DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&dcr); err != nil {
		return nil, fmt.Errorf("decoding device code response: %w", err)
	}

	return &dcr, nil
}

// PollForCredentials polls the device credentials endpoint until authorization
// is granted, the timeout is reached, or an unrecoverable error occurs.
func PollForCredentials(baseURL, deviceCode, clientID string, interval int, timeout time.Duration) (*DeviceCredentials, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := time.Duration(interval) * time.Second

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("device authorization timed out")
		}

		reqURL := fmt.Sprintf("%s/device/credentials?client_id=%s&code=%s",
			baseURL, url.QueryEscape(clientID), url.QueryEscape(deviceCode))

		resp, err := http.Get(reqURL)
		if err != nil {
			return nil, fmt.Errorf("polling for credentials: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			var creds DeviceCredentials
			if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("decoding credentials: %w", err)
			}
			resp.Body.Close()
			return &creds, nil
		}

		resp.Body.Close()

		if resp.StatusCode == http.StatusForbidden {
			time.Sleep(pollInterval)
			continue
		}

		return nil, fmt.Errorf("unexpected status polling credentials: HTTP %d", resp.StatusCode)
	}
}

// ExchangeToken exchanges a device code for an access token.
func ExchangeToken(baseURL, clientID, clientSecret, deviceCode string) (*TokenData, error) {
	form := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {deviceCode},
		"grant_type":    {"http://oauth.net/grant_type/device/1.0"},
	}

	resp, err := http.PostForm(baseURL+"/token", form)
	if err != nil {
		return nil, fmt.Errorf("exchanging token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokens TokenData
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	tokens.ClientID = clientID
	tokens.ClientSecret = clientSecret
	tokens.ObtainedAt = time.Now().Unix()

	return &tokens, nil
}

// RefreshAccessToken uses the refresh token to obtain a new access token.
func RefreshAccessToken(baseURL string, tokens *TokenData) (*TokenData, error) {
	form := url.Values{
		"client_id":     {tokens.ClientID},
		"client_secret": {tokens.ClientSecret},
		"code":          {tokens.RefreshToken},
		"grant_type":    {"http://oauth.net/grant_type/device/1.0"},
	}

	resp, err := http.PostForm(baseURL+"/token", form)
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var newTokens TokenData
	if err := json.NewDecoder(resp.Body).Decode(&newTokens); err != nil {
		return nil, fmt.Errorf("decoding refresh response: %w", err)
	}

	newTokens.ClientID = tokens.ClientID
	newTokens.ClientSecret = tokens.ClientSecret
	newTokens.ObtainedAt = time.Now().Unix()

	return &newTokens, nil
}
