package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [path_or_id]",
	Short: "Delete file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		
		obj, err := findObjectByPath(driver, args[0])
		if err != nil {
			return err
		}
		
		if err := driver.Delete(context.Background(), obj); err != nil {
			return err
		}
		
		fmt.Printf("Deleted %s\n", obj.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
