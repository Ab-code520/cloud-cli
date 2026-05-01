package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cloud-cli/core"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	driverName string
	jsonOutput bool
	threads    int
)

var rootCmd = &cobra.Command{
	Use:   "cloud-cli",
	Short: "高扩展微内核网盘 CLI 工具",
	Long: `cloud-cli 是一个高扩展、微内核架构的网盘 CLI 工具。
支持夸克、百度等多种网盘，通过统一的 Storage 接口实现。

用法示例:
  cloud-cli quark ls /
  cloud-cli quark upload local.txt /remote/dir/
  cloud-cli quark download /remote/file.txt ./local.txt`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cfgFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				home = "."
			}
			cfgFile = filepath.Join(home, ".config", "cloud-cli", "config.json")
		}
		if err := core.InitConfig(cfgFile); err != nil {
			return fmt.Errorf("init config: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().IntVarP(&threads, "t", "t", 3, "max concurrent threads")
}

func GetDriver() (core.Storage, error) {
	if driverName == "" {
		driverName = "quark"
	}
	driver, err := core.NewDriver(driverName)
	if err != nil {
		return nil, err
	}
	tokens, err := core.GetDriverTokens(driverName)
	if err != nil {
		return nil, fmt.Errorf("driver '%s' not configured, use 'login' command first", driverName)
	}
	if err := driver.Init(tokens); err != nil {
		return nil, fmt.Errorf("init driver: %w", err)
	}
	return driver, nil
}
