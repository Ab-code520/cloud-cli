package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [file_id]",
	Short: "Show detailed information about a file or folder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileID := args[0]
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		obj, err := driver.Info(cmd.Context(), fileID)
		if err != nil {
			return err
		}

		return outputTableOrJSON(obj, func() {
			typeStr := "File"
			if obj.IsDir {
				typeStr = "Folder"
			}

			fmt.Printf("📂 Name: %s\n", obj.Name)
			fmt.Printf("🆔 ID: %s\n", obj.ID)
			fmt.Printf("📦 Type: %s\n", typeStr)
			
			if !obj.IsDir {
				fmt.Printf("💾 Size: %s\n", formatSize(obj.Size))
				if obj.Hash != "" {
					fmt.Printf("#️⃣ SHA1: %s\n", obj.Hash)
				}
			}
			
			fmt.Printf("🕒 Modified: %s\n", obj.ModTime.Format("2006-01-02 15:04:05"))
		})
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
