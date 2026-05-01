package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Ab-code520/cloud-cli/backends/quark"
	"github.com/Ab-code520/cloud-cli/core"
	"github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [driver]",
	Short: "Login to cloud drive (supports QR scan or manual cookie)",
	Long: `Login to a cloud drive backend.

Supported methods:
  1. QR Scan (interactive): cloud-cli login quark
  2. Manual Cookie:         cloud-cli login quark --cookie "your_cookie_string"

Supported drivers: quark`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driverName := args[0]

		cookie, err := cmd.Flags().GetString("cookie")
		if err != nil {
			return err
		}

		// Mode 1: QR Scan login (no cookie provided)
		if cookie == "" {
			return loginWithQR(cmd.Context(), driverName)
		}

		// Mode 2: Manual cookie login
		return loginWithCookie(driverName, cookie)
	},
}

// loginWithQR performs QR code based login.
func loginWithQR(ctx context.Context, driverName string) error {
	if driverName != "quark" {
		return fmt.Errorf("QR login only supports 'quark' driver currently")
	}

	// Setup context with timeout (5 minutes)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Handle Ctrl+C gracefully
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("🔐 Initializing QR login for Quark Drive...")

	api := quark.NewAPI("")
	qrResp, err := api.GenerateQR(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	fmt.Println("\n📱 Please scan the QR code with Quark app:")
	fmt.Println("┌" + strings.Repeat("─", 39) + "┐")
	
	// Render QR code in terminal
	qrterminal.GenerateHalfBlock(qrResp.URL, qrterminal.L, os.Stdout)

	fmt.Println("└" + strings.Repeat("─", 39) + "┘")
	fmt.Println("\nWaiting for scan confirmation... (timeout: 5 minutes)")

	// Poll QR status
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	attempts := 0
	maxAttempts := 150 // 5 minutes / 2 seconds

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("login cancelled: %w", ctx.Err())
		case <-ticker.C:
			attempts++
			if attempts >= maxAttempts {
				return fmt.Errorf("login timeout - QR code expired")
			}

			status, err := api.QueryQRStatus(ctx, qrResp.Token)
			if err != nil {
				fmt.Printf("\r⚠️  Polling error: %v", err)
				continue
			}

			switch status.Status {
			case quark.QRNotScanned:
				fmt.Print("\r⏳ Waiting for scan...")
			case quark.QRScanned:
				fmt.Print("\r✅ QR scanned! Confirming login...")
			case quark.QRConfirmed:
				fmt.Println("\n🎉 Login confirmed!")

				// Extract cookie from status response
				cookie := extractCookie(status)
				if cookie == "" {
					return fmt.Errorf("login confirmed but no cookie received")
				}

				return loginWithCookie(driverName, cookie)
			case quark.QRExpired:
				fmt.Println("\n❌ QR code expired. Please try again.")
				return fmt.Errorf("QR code expired")
			default:
				fmt.Printf("\r⚠️  Unknown status: %s", status.Status)
			}
		}
	}
}

// extractCookie extracts cookie from QR query response.
// Quark may return cookie in different fields depending on the API version.
func extractCookie(status *quark.QRQueryResp) string {
	// Direct cookie field
	if status.Cookie != "" {
		return status.Cookie
	}
	// If redirect_url contains ticket, we may need to extract it
	// For now, return what we have
	return ""
}

// loginWithCookie saves the cookie to config file.
func loginWithCookie(driverName string, cookie string) error {
	// Load or initialize config
	cfg, err := core.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	core.GlobalConfig = cfg

	// Save account
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

	// Set as default
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
	loginCmd.Flags().StringP("cookie", "c", "", "Cookie string (skip QR login)")
	rootCmd.AddCommand(loginCmd)
}

// Helper for string repeat (Go 1.21+)
type stringRepeat int

func (n stringRepeat) repeat() string {
	return ""
}
