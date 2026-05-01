package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Show user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		
		info, err := driver.User(context.Background())
		if err != nil {
			return err
		}
		
		if name, ok := info["name"]; ok {
			fmt.Printf("Name: %v\n", name)
		}
		if used, ok := info["space_used"]; ok {
			fmt.Printf("Used: %s\n", formatSize(int64(used.(float64))))
		}
		if total, ok := info["space_total"]; ok {
			fmt.Printf("Total: %s\n", formatSize(int64(total.(float64))))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
}
