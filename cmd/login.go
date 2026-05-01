package cmd

import (
	"fmt"
	"os"

	"cloud-cli/core"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [driver]",
	Short: "登录网盘账号",
	Long: `登录网盘并保存认证信息。
示例:
  cloud-cli login quark --cookie "your_cookie_here"
  cloud-cli login quark --env KUAKE_COOKIE`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driverName = args[0]

		cookie, _ := cmd.Flags().GetString("cookie")
		envVar, _ := cmd.Flags().GetString("env")
		
		if envVar != "" {
			cookie = os.Getenv(envVar)
			if cookie == "" {
				return fmt.Errorf("environment variable %s is not set", envVar)
			}
		}

		if cookie == "" {
			return fmt.Errorf("please provide --cookie or --env flag")
		}

		_, err := core.NewDriver(driverName)
		if err != nil {
			return err
		}

		tokens := map[string]string{
			"cookie": cookie,
		}

		if err := core.SetDriverTokens(driverName, tokens); err != nil {
			return err
		}

		fmt.Printf("✅ %s 登录成功，配置已保存至 %s\n", driverName, cfgFile)
		return nil
	},
}

func init() {
	loginCmd.Flags().String("cookie", "", "网盘 Cookie 值")
	loginCmd.Flags().String("env", "", "从环境变量读取 Cookie (如: KUAKE_COOKIE)")
	rootCmd.AddCommand(loginCmd)
}
