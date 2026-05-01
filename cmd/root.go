package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cloud-cli",
	Short: "Universal Cloud Drive CLI",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getDriver() (core.Storage, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "cloud-cli", "config.json")
	cfg, err := core.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	
	driverName := cfg.CurrentDriver
	if driverName == "" {
		return nil, fmt.Errorf("no driver selected, use 'login' first")
	}
	
	driverConfig, ok := cfg.GetDriverConfig(driverName)
	if !ok {
		return nil, fmt.Errorf("driver %s not configured", driverName)
	}
	
	driver, err := core.NewDriver(driverName)
	if err != nil {
		return nil, err
	}
	
	if err := driver.Init(driverConfig.Tokens); err != nil {
		return nil, err
	}
	
	return driver, nil
}
