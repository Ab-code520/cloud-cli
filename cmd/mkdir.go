package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir [path]",
	Short: "Create directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		
		_, err = driver.Mkdir(context.Background(), args[0])
		if err != nil {
			return err
		}
		
		fmt.Printf("Created directory %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkdirCmd)
}
