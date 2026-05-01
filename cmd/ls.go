package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		
		path := "/"
		if len(args) > 0 {
			path = args[0]
		}
		
		objects, err := driver.List(context.Background(), path)
		if err != nil {
			return err
		}
		
		fmt.Printf("%-30s %-10s %s\n", "NAME", "SIZE", "TIME")
		for _, obj := range objects {
			size := "-"
			if !obj.IsDir {
				size = formatSize(obj.Size)
			} else {
				size = "DIR"
			}
			name := obj.Name
			if obj.IsDir {
				name += "/"
			}
			fmt.Printf("%-30s %-10s %s\n", name, size, obj.ModTime.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
