package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [local_path...] [remote_dir_id]",
	Short: "Upload file(s) to cloud drive",
	Long: `Upload file(s) to cloud drive with optional policy control.

Examples:
  cloud-cli upload file.txt              # Upload to root
  cloud-cli upload file.txt folder123    # Upload to specific folder
  cloud-cli upload *.mp4 folder123       # Batch upload
  cloud-cli upload file.txt --policy skip     # Skip if exists
  cloud-cli upload file.txt --policy overwrite # Overwrite if exists
  cloud-cli upload file.txt --policy rsync   # Upload only if modified`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		policy, _ := cmd.Flags().GetString("policy")
		if policy != "skip" && policy != "overwrite" && policy != "rsync" {
			policy = "skip" // Default policy
		}

		// Last arg is remote dir ID if it looks like an ID, otherwise root
		remoteDirID := "0"
		localPaths := args

		// Check if last arg is a directory ID (alphanumeric)
		if len(args) > 1 {
			lastArg := args[len(args)-1]
			if isValidDirID(lastArg) {
				remoteDirID = lastArg
				localPaths = args[:len(args)-1]
			}
		}

		remoteDir := &core.Object{ID: remoteDirID, Name: "/", IsDir: true}

		uploaded := 0
		skipped := 0
		failed := 0

		for _, localPath := range localPaths {
			// Handle glob patterns
			matches, err := filepath.Glob(localPath)
			if err != nil {
				fmt.Printf("❌ Invalid pattern '%s': %v\n", localPath, err)
				failed++
				continue
			}
			if len(matches) == 0 {
				matches = []string{localPath}
			}

			for _, path := range matches {
				info, err := os.Stat(path)
				if err != nil {
					fmt.Printf("❌ Cannot access '%s': %v\n", path, err)
					failed++
					continue
				}

				if info.IsDir() {
					fmt.Printf("⚠️ Skipping directory '%s' (use recursive upload flag in future)\n", path)
					skipped++
					continue
				}

				// Check policy
				if policy == "skip" {
					exists, err := fileExistsOnRemote(driver, cmd.Context(), remoteDirID, info.Name())
					if err == nil && exists {
						fmt.Printf("⏭️ Skipped '%s' (already exists)\n", info.Name())
						skipped++
						continue
					}
				}

				f, err := os.Open(path)
				if err != nil {
					fmt.Printf("❌ Cannot open '%s': %v\n", path, err)
					failed++
					continue
				}

				fmt.Printf("⬆️ Uploading %s (%s)...\n", info.Name(), formatSize(info.Size()))

				_, err = driver.Put(context.Background(), f, remoteDir, info.Name(), info.Size())
				f.Close()

				if err != nil {
					fmt.Printf("❌ Failed to upload '%s': %v\n", info.Name(), err)
					failed++
				} else {
					fmt.Printf("✅ Uploaded: %s\n", info.Name())
					uploaded++
				}
			}
		}

		fmt.Printf("\n📊 Summary: %d uploaded, %d skipped, %d failed\n", uploaded, skipped, failed)
		return nil
	},
}

// isValidDirID checks if a string looks like a directory ID (alphanumeric, length 32+)
func isValidDirID(s string) bool {
	if len(s) < 10 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// fileExistsOnRemote checks if a file with the same name exists in the remote directory
func fileExistsOnRemote(driver core.Storage, ctx context.Context, dirID, fileName string) (bool, error) {
	objects, err := driver.List(ctx, dirID)
	if err != nil {
		return false, err
	}
	for _, obj := range objects {
		if obj.Name == fileName && !obj.IsDir {
			return true, nil
		}
	}
	return false, nil
}

func init() {
	uploadCmd.Flags().StringP("policy", "p", "skip", "Upload policy: skip|overwrite|rsync")
	rootCmd.AddCommand(uploadCmd)
}
