package cmd

import (
	"fmt"
	"os"

	"cloud-cli/core"
	"cloud-cli/utils"

	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <local> <remote>",
	Short: "上传文件到网盘",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := GetDriver()
		if err != nil {
			return err
		}

		localPath := args[0]
		remoteDir := args[1]

		info, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("local file not found: %w", err)
		}
		if info.IsDir() {
			return fmt.Errorf("directory upload not yet supported")
		}

		remoteObj := &core.Object{Path: remoteDir}

		fmt.Printf("上传中: %s -> %s\n", localPath, remoteDir)
		pb := utils.NewProgressBar(info.Size(), "B", !jsonOutput)

		err = driver.Upload(localPath, remoteObj, threads)
		if err != nil {
			return err
		}

		pb.Done()
		fmt.Printf("✅ 上传完成: %s (%s)\n", info.Name(), utils.FormatSize(info.Size()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
