package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/joeyfurness/rdl/internal/api"
	"github.com/joeyfurness/rdl/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Real-Debrid authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Real-Debrid using device flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doDeviceAuth()
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.AuthPath()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println("Not authenticated — nothing to do.")
			return nil
		}
		if err := api.DeleteTokens(path); err != nil {
			return fmt.Errorf("deleting tokens: %w", err)
		}
		fmt.Println("Logged out successfully.")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication state",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check env var first
		if token := os.Getenv("RDL_TOKEN"); token != "" {
			fmt.Println("Authenticated via RDL_TOKEN environment variable.")
			return nil
		}

		tokens, err := api.LoadTokens(config.AuthPath())
		if err != nil {
			fmt.Println("Not authenticated.")
			return nil
		}

		if tokens.IsExpired() {
			fmt.Println("Authenticated (token expired — will refresh on next use).")
		} else {
			fmt.Println("Authenticated.")
		}
		return nil
	},
}

// doDeviceAuth runs the full OAuth device-code flow.
func doDeviceAuth() error {
	dcr, err := api.RequestDeviceCode(api.OAuthBaseURL, api.DefaultClientID)
	if err != nil {
		return fmt.Errorf("requesting device code: %w", err)
	}

	fmt.Printf("Go to: %s\n", dcr.VerificationURL)
	fmt.Printf("Enter code: %s\n", dcr.UserCode)
	fmt.Println("Waiting for authorization...")

	timeout := time.Duration(dcr.ExpiresIn) * time.Second
	creds, err := api.PollForCredentials(api.OAuthBaseURL, dcr.DeviceCode, api.DefaultClientID, dcr.Interval, timeout)
	if err != nil {
		return fmt.Errorf("polling for credentials: %w", err)
	}

	tokens, err := api.ExchangeToken(api.OAuthBaseURL, creds.ClientID, creds.ClientSecret, dcr.DeviceCode)
	if err != nil {
		return fmt.Errorf("exchanging token: %w", err)
	}

	if err := api.SaveTokens(config.AuthPath(), tokens); err != nil {
		return fmt.Errorf("saving tokens: %w", err)
	}

	// Verify by fetching user info
	client := api.NewClient(tokens.AccessToken)
	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Type     string `json:"type"`
	}
	if err := client.Get("/user", &user); err != nil {
		fmt.Println("Tokens saved but failed to verify user:", err)
		return nil
	}

	fmt.Printf("Authenticated as %s (%s)\n", user.Username, user.Email)
	return nil
}

// GetAuthenticatedClient returns an authenticated API client.
// It checks RDL_TOKEN env var first, then saved tokens, and falls back
// to triggering the device auth flow if needed.
func GetAuthenticatedClient() (*api.Client, error) {
	// 1. Check env var
	if token := os.Getenv("RDL_TOKEN"); token != "" {
		return api.NewClient(token), nil
	}

	// 2. Try loading saved tokens
	tokens, err := api.LoadTokens(config.AuthPath())
	if err != nil {
		// No tokens — trigger auth flow
		fmt.Println("No credentials found. Starting authentication...")
		if err := doDeviceAuth(); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		tokens, err = api.LoadTokens(config.AuthPath())
		if err != nil {
			return nil, fmt.Errorf("loading tokens after auth: %w", err)
		}
		return api.NewClient(tokens.AccessToken), nil
	}

	// 3. If expired, try refresh
	if tokens.IsExpired() {
		newTokens, err := api.RefreshAccessToken(api.OAuthBaseURL, tokens)
		if err != nil {
			// Refresh failed — re-auth
			fmt.Println("Token refresh failed. Re-authenticating...")
			if err := doDeviceAuth(); err != nil {
				return nil, fmt.Errorf("re-authentication failed: %w", err)
			}
			tokens, err = api.LoadTokens(config.AuthPath())
			if err != nil {
				return nil, fmt.Errorf("loading tokens after re-auth: %w", err)
			}
			return api.NewClient(tokens.AccessToken), nil
		}
		if err := api.SaveTokens(config.AuthPath(), newTokens); err != nil {
			return nil, fmt.Errorf("saving refreshed tokens: %w", err)
		}
		return api.NewClient(newTokens.AccessToken), nil
	}

	return api.NewClient(tokens.AccessToken), nil
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
}
