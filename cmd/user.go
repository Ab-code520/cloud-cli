package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "查看用户信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := GetDriver()
		if err != nil {
			return err
		}

		info, err := driver.User()
		if err != nil {
			return err
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Try to extract common fields
		if data, ok := info["data"].(map[string]interface{}); ok {
			if name, ok := data["nickname"].(string); ok {
				fmt.Printf("昵称: %s\n", name)
			}
			if cap, ok := data["total_size"].(float64); ok {
				fmt.Printf("总容量: %s\n", formatSize(int64(cap)))
			}
			if used, ok := data["used_size"].(float64); ok {
				fmt.Printf("已使用: %s\n", formatSize(int64(used)))
			}
		} else {
			fmt.Printf("用户信息: %+v\n", info)
		}

		return nil
	},
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024*1024*1024:
		return fmt.Sprintf("%.2f TB", float64(bytes)/1024/1024/1024/1024)
	case bytes >= 1024*1024*1024:
		return fmt.Sprintf("%.2f GB", float64(bytes)/1024/1024/1024)
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func init() {
	rootCmd.AddCommand(userCmd)
}
