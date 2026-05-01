package cmd
import (
	"fmt"
	"os"
	"strings"

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

func initDriver(name string, rootPath string) (core.Storage, error) {
	if core.GlobalConfig == nil {
		cfg, err := core.LoadConfig("")
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
		core.GlobalConfig = cfg
	}

	driver, err := core.NewDriver(name)
	if err != nil {
		return nil, fmt.Errorf("unsupported driver '%s': %w", name, err)
	}

	var initCfg map[string]string

	if name == "local" {
		initCfg = map[string]string{
			"root": rootPath,
		}
	} else {
		// Load from config
		// Try specific account name if provided, else default
		// For simplicity, let's look for an account of this type
		found := false
		for _, accName := range core.GlobalConfig.ListAccounts() {
			acc, _ := core.GlobalConfig.GetAccount(accName)
			if acc.Type == name {
				initCfg = make(map[string]string)
				for k, v := range acc.Cookie {
					initCfg[k] = v
				}
				for k, v := range acc.Params {
					initCfg[k] = v
				}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("no account found for driver '%s'. Please login first.", name)
		}
	}

	if err := driver.Init(initCfg); err != nil {
		return nil, fmt.Errorf("init driver %s: %w", name, err)
	}

	return driver, nil
}

// parseResourcePath parses "driver:path" into (driver, path).
// If no driver prefix, assumes "local".
func parseResourcePath(input string) (string, string, error) {
	// Split by first ":"
	idx := strings.Index(input, ":")
	if idx == -1 {
		// No driver specified -> local
		return "local", input, nil
	}
	
	driver := input[:idx]
	path := input[idx+1:]
	
	// Handle Windows paths like C:\... -> driver "C", path "\..."
	if len(driver) == 1 && driver[0] >= 'A' && driver[0] <= 'Z' {
		// Probably a Windows drive letter, treat as local
		return "local", input, nil
	}
	if len(driver) == 1 && driver[0] >= 'a' && driver[0] <= 'z' {
		 return "local", input, nil
	}

	return driver, path, nil
}

func getDriver() (core.Storage, error) {
	if core.GlobalConfig == nil {
		cfg, err := core.LoadConfig("")
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
		core.GlobalConfig = cfg
	}

	acc, err := core.GlobalConfig.GetDefaultAccount()
	if err != nil {
		return nil, fmt.Errorf("no default account selected, use 'login' or 'account switch' first")
	}

	return initDriver(acc.Type, "")
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json")
}
