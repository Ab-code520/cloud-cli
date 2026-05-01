package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <path>",
	Short: "创建目录",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := GetDriver()
		if err != nil {
			return err
		}

		path := args[0]
		err = driver.Mkdir(path)
		if err != nil {
			return err
		}

		fmt.Printf("✅ 目录已创建: %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkdirCmd)
}
