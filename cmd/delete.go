package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id_or_path...]",
	Short: "Delete file(s) or folder(s)",
	Long: `Delete one or more files or folders.

Examples:
  cloud-cli delete file_id                # Delete single file
  cloud-cli delete id1 id2 id3            # Batch delete
  cloud-cli delete /path/to/file.txt      # Delete by path`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		deleted := 0
		failed := 0

		for _, arg := range args {
			obj, err := findObjectByPath(driver, arg)
			if err != nil {
				fmt.Printf("❌ Cannot find '%s': %v\n", arg, err)
				failed++
				continue
			}

			if err := driver.Delete(context.Background(), obj); err != nil {
				fmt.Printf("❌ Failed to delete '%s': %v\n", obj.Name, err)
				failed++
			} else {
				fmt.Printf("🗑️ Deleted: %s\n", obj.Name)
				deleted++
			}
		}

		fmt.Printf("\n📊 Summary: %d deleted, %d failed\n", deleted, failed)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
