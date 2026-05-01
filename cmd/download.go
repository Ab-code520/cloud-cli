package cmd

import (
	"fmt"
	"strings"

	"cloud-cli/core"
	"cloud-cli/utils"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <remote> <local>",
	Short: "从网盘下载文件",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := GetDriver()
		if err != nil {
			return err
		}

		remotePath := args[0]
		localPath := args[1]

		// 获取文件所在目录和文件名
		dirPath, fileName := splitPath(remotePath)
		
		// 列出目录找到文件
		objects, err := driver.List(dirPath)
		if err != nil {
			return err
		}

		var targetObj *core.Object
		for _, obj := range objects {
			if obj.Name == fileName && !obj.IsDir {
				targetObj = obj
				break
			}
		}

		if targetObj == nil {
			return fmt.Errorf("file not found: %s", remotePath)
		}

		fmt.Printf("下载中: %s (%s) -> %s\n", targetObj.Name, utils.FormatSize(targetObj.Size), localPath)

		err = driver.Download(targetObj, localPath, threads)
		if err != nil {
			return err
		}

		fmt.Printf("✅ 下载完成: %s (%s)\n", targetObj.Name, utils.FormatSize(targetObj.Size))
		return nil
	},
}

// splitPath 分割路径为目录和文件名
func splitPath(path string) (string, string) {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return "/", ""
	}

	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return "/", path
	}

	return "/" + path[:lastSlash], path[lastSlash+1:]
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
