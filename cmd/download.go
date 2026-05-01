package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [remote_path] [local_path]",
	Short: "Download file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		
		remotePath := args[0]
		localPath := args[1]
		
		// Fix #8: Support recursive path finding
		obj, err := findObjectByPathOrID(driver, remotePath)
		if err != nil {
			return err
		}
		
		if obj.IsDir {
			return fmt.Errorf("download does not support directories yet")
		}
		
		// Fix #7: Use cmd.Context() for cancellation
		reader, err := driver.Open(cmd.Context(), obj, 0)
		if err != nil {
			return err
		}
		defer reader.Close()
		
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return err
		}
		
		out, err := os.Create(localPath)
		if err != nil {
			return err
		}
		defer out.Close()
		
		fmt.Printf("Downloading %s (%s)...\n", obj.Name, formatSize(obj.Size))
		
		// Simple copy, progress bar could be added here via io.TeeReader
		_, err = io.Copy(out, reader)
		if err != nil {
			return err
		}
		
		fmt.Printf("Downloaded to %s\n", localPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
