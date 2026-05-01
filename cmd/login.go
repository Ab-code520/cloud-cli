package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [driver]",
	Short: "Login to cloud drive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driverName := args[0]
		
		cookie, err := cmd.Flags().GetString("cookie")
		if err != nil {
			return err
		}
		
		if cookie == "" {
			fmt.Print("Enter Cookie: ")
			fmt.Scanln(&cookie)
		}
		
		cfg := core.Config{
			CurrentDriver: driverName,
			Drivers: map[string]*core.DriverConfig{
				driverName: {
					Type: driverName,
					Tokens: map[string]string{
						"cookie": cookie,
					},
				},
			},
		}
		
		configPath := filepath.Join(os.Getenv("HOME"), ".config", "cloud-cli", "config.json")
		if err := core.SaveConfig(configPath, &cfg); err != nil {
			return err
		}
		
		fmt.Printf("Logged in to %s successfully.\n", driverName)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringP("cookie", "c", "", "Cookie string")
	rootCmd.AddCommand(loginCmd)
}
