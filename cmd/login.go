package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [driver] --cookie \"cookie_string\"",
	Short: "Login to cloud drive by providing a cookie string",
	Long: `Login to a cloud drive backend by manually providing a cookie.

Usage:
  cloud-cli login quark --cookie "your_cookie_string"

Supported drivers: quark`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driverName := args[0]

		cookie, err := cmd.Flags().GetString("cookie")
		if err != nil {
			return err
		}
		if cookie == "" {
			return fmt.Errorf("cookie is required. Use --cookie flag to provide your login cookie.\nExample: cloud-cli login quark --cookie \"your_cookie_string\"")
		}

		return loginWithCookie(driverName, cookie)
	},
}

// loginWithCookie saves the cookie to config file.
func loginWithCookie(driverName string, cookie string) error {
	cfg, err := core.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	core.GlobalConfig = cfg

	acc := &core.Account{
		Type: driverName,
		Cookie: map[string]string{
			"cookie": cookie,
		},
	}

	accountName := driverName + "-default"
	if err := core.GlobalConfig.AddAccount(accountName, acc); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	core.GlobalConfig.Default = accountName
	if err := core.GlobalConfig.Save(); err != nil {
		return fmt.Errorf("failed to update default: %w", err)
	}

	configPath := filepath.Join(os.Getenv("HOME"), ".config", "cloud-cli", "config.yaml")
	fmt.Printf("✅ Logged in to %s successfully.\n", driverName)
	fmt.Printf("📁 Config saved to: %s\n", configPath)
	return nil
}

func init() {
	loginCmd.Flags().StringP("cookie", "c", "", "Cookie string for authentication")
	rootCmd.AddCommand(loginCmd)
}
