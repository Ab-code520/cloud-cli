package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share [subcommand]",
	Short: "Manage file shares (create, list, delete)",
	Long: `Manage file shares for cloud drive.

Examples:
  cloud-cli share create file_id1 file_id2    # Create a share for files
  cloud-cli share list                        # List all shares
  cloud-cli share delete share_id1 share_id2  # Delete shares`,
}

var shareCreateCmd = &cobra.Command{
	Use:   "create [file_id...]",
	Short: "Create a share for files/folders",
	Long: `Create a share for one or more files or folders.

Examples:
  cloud-cli share create abc123              # Permanent share
  cloud-cli share create abc123 --days 7     # 7-day share`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		days, _ := cmd.Flags().GetInt("days")
		share, err := driver.CreateShare(cmd.Context(), args, days)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Share created successfully!\n")
		fmt.Printf("🔗 URL: %s\n", share.URL)
		fmt.Printf("🆔 Share ID: %s\n", share.ShareID)
		return nil
	},
}

var shareListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all shares",
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		shares, err := driver.ListShares(cmd.Context(), 1, 50)
		if err != nil {
			return err
		}

		if len(shares) == 0 {
			fmt.Println("No shares found.")
			return nil
		}

		fmt.Printf("📦 Found %d shares:\n\n", len(shares))
		for _, s := range shares {
			status := "🟢 Active"
			if s.IsExpired {
				status = "🔴 Expired"
			}
			fmt.Printf("%s %s\n", status, s.Title)
			fmt.Printf("   ID: %s\n", s.ShareID)
			fmt.Printf("   🔗 %s\n", s.URL)
			fmt.Printf("   📄 Files: %d | 📅 Created: %s\n\n", s.FileCount, time.Unix(s.CreatedAt, 0).Format("2006-01-02"))
		}
		return nil
	},
}

var shareDeleteCmd = &cobra.Command{
	Use:   "delete [share_id...]",
	Short: "Delete one or more shares",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		err = driver.DeleteShare(cmd.Context(), args)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Deleted %d share(s) successfully.\n", len(args))
		return nil
	},
}

func init() {
	shareCreateCmd.Flags().IntP("days", "d", -1, "Expiration days (-1 for permanent)")
	
	shareCmd.AddCommand(shareCreateCmd)
	shareCmd.AddCommand(shareListCmd)
	shareCmd.AddCommand(shareDeleteCmd)
	rootCmd.AddCommand(shareCmd)
}
