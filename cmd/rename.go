package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename [file_id] [new_name]",
	Short: "Rename a file or folder",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fileID := args[0]
		newName := args[1]
		
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		obj, err := driver.Info(cmd.Context(), fileID)
		if err != nil {
			return fmt.Errorf("get file info: %w", err)
		}

		err = driver.Rename(cmd.Context(), obj, newName)
		if err != nil {
			return fmt.Errorf("rename failed: %w", err)
		}

		fmt.Printf("✅ Successfully renamed '%s' to '%s'\n", obj.Name, newName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
