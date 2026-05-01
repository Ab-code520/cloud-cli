package cmd

import (
	"fmt"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy [source_id] [dest_dir_id]",
	Short: "Copy a file or folder to another directory",
	Long: `Copy a file or folder to another directory.

Examples:
  cloud-cli copy file123 folder456   # Copy file123 into folder456
  cloud-cli copy file123 0           # Copy file123 to root directory`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcID := args[0]
		destID := args[1]
		
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		// Get source object
		srcObj, err := driver.Info(cmd.Context(), srcID)
		if err != nil {
			return fmt.Errorf("get source info: %w", err)
		}

		// Construct destination directory object
		destDirObj := &core.Object{ID: destID}

		// Call Copy
		err = driver.Copy(cmd.Context(), srcObj, destDirObj)
		if err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}

		fmt.Printf("✅ Successfully copied '%s' to directory '%s'\n", srcObj.Name, destID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
