package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "rm [path...]",
	Aliases: []string{"delete"},
	Short:   "删除文件或目录",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := GetDriver()
		if err != nil {
			return err
		}

		for _, path := range args {
			objects, err := driver.List(path)
			if err != nil {
				fmt.Printf("⚠️  删除 %s 失败: %v\n", path, err)
				continue
			}

			for _, obj := range objects {
				err := driver.Delete(obj)
				if err != nil {
					fmt.Printf("⚠️  删除 %s 失败: %v\n", obj.Name, err)
				} else {
					fmt.Printf("✅ 已删除: %s\n", obj.Name)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
