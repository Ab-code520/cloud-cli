package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [local_path] [remote_dir]",
	Short: "Upload file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		
		localPath := args[0]
		_ = args[1] // remoteDirPath reserved for future traversal
		
		info, err := os.Stat(localPath)
		if err != nil {
			return err
		}
		
		f, err := os.Open(localPath)
		if err != nil {
			return err
		}
		defer f.Close()
		
		remoteDir := &core.Object{ID: "0", Name: "/", IsDir: true}
		
		fmt.Printf("Uploading %s (%s)...\n", info.Name(), formatSize(info.Size()))
		
		obj, err := driver.Put(context.Background(), f, remoteDir, info.Name(), info.Size())
		if err != nil {
			return err
		}
		
		fmt.Printf("Upload complete: %s (ID: %s)\n", obj.Name, obj.ID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
