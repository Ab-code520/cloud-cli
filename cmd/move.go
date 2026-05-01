package cmd

import (
	"fmt"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move [source_id] [dest_dir_id]",
	Short: "Move a file or folder to another directory",
	Long: `Move a file or folder to another directory.

Examples:
  cloud-cli move file123 folder456   # Move file123 into folder456
  cloud-cli move file123 0           # Move file123 to root directory`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcID := args[0]
		destID := args[1]
		
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		srcObj, err := driver.Info(cmd.Context(), srcID)
		if err != nil {
			return fmt.Errorf("get source info: %w", err)
		}

		destDirObj := &core.Object{ID: destID}
		err = driver.Move(cmd.Context(), srcObj, destDirObj)
		if err != nil {
			return fmt.Errorf("move failed: %w", err)
		}

		fmt.Printf("✅ Successfully moved '%s' to directory '%s'\n", srcObj.Name, destID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
