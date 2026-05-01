package cmd

import (
	"encoding/json"
	"fmt"

	"cloud-cli/utils"

	"github.com/spf13/cobra"
)

// LSObject 用于 JSON 输出的简化对象
type LSObject struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	SizeFmt string `json:"size_formatted"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
	Path    string `json:"path"`
}

var lsCmd = &cobra.Command{
	Use:     "ls [path]",
	Aliases: []string{"list"},
	Short:   "列出目录内容",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := GetDriver()
		if err != nil {
			return err
		}

		path := "/"
		if len(args) > 0 {
			path = args[0]
		}

		objects, err := driver.List(path)
		if err != nil {
			return err
		}

		if jsonOutput {
			var result []LSObject
			for _, obj := range objects {
				result = append(result, LSObject{
					ID:      obj.ID,
					Name:    obj.Name,
					Size:    obj.Size,
					SizeFmt: utils.FormatSize(obj.Size),
					IsDir:   obj.IsDir,
					ModTime: obj.ModTime.Format("2006-01-02 15:04:05"),
					Path:    obj.Path,
				})
			}
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("json marshal: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		if len(objects) == 0 {
			fmt.Println("(empty directory)")
			return nil
		}

		for _, obj := range objects {
			icon := "📄"
			if obj.IsDir {
				icon = "📁"
			}
			sizeStr := "-"
			if !obj.IsDir {
				sizeStr = utils.FormatSize(obj.Size)
			}
			fmt.Printf("%s  %-50s  %10s  %s\n", icon, obj.Name, sizeStr, obj.ModTime.Format("2006-01-02 15:04"))
		}
		fmt.Printf("\n共 %d 个项目\n", len(objects))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
