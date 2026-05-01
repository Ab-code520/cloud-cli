package cmd

import (
	"fmt"
	"time"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var recycleCmd = &cobra.Command{
	Use:   "recycle [subcommand]",
	Short: "Manage recycle bin (list, restore, delete)",
	Long: `Manage recycle bin items.

Examples:
  cloud-cli recycle list                    # List deleted files
  cloud-cli recycle restore fid1 fid2       # Restore files
  cloud-cli recycle delete fid1 fid2        # Permanently delete`,
}

var recycleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List items in recycle bin",
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		rb, ok := driver.(core.RecycleBin)
		if !ok {
			return fmt.Errorf("recycle bin is not supported by this driver")
		}

		page, _ := cmd.Flags().GetInt("page")
		size, _ := cmd.Flags().GetInt("size")

		items, err := rb.ListRecycle(cmd.Context(), page, size)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			fmt.Println("♻️ Recycle bin is empty.")
			return nil
		}

		fmt.Printf("♻️ Found %d items in recycle bin:\n\n", len(items))
		for _, item := range items {
			typeIcon := "📄"
			if item.IsDir {
				typeIcon = "📁"
			}
			sizeStr := ""
			if !item.IsDir {
				sizeStr = fmt.Sprintf(" (%s)", formatSize(item.Size))
			}
			fmt.Printf("%s %s%s\n", typeIcon, item.FileName, sizeStr)
			fmt.Printf("   🆔 %s | 🗑️ Deleted: %s\n\n", item.FID, time.Unix(item.DeletedAt, 0).Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

var recycleRestoreCmd = &cobra.Command{
	Use:   "restore [file_id...]",
	Short: "Restore items from recycle bin",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		rb, ok := driver.(core.RecycleBin)
		if !ok {
			return fmt.Errorf("recycle bin is not supported by this driver")
		}

		err = rb.RecoverRecycle(cmd.Context(), args)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Successfully restored %d item(s).\n", len(args))
		return nil
	},
}

var recycleDeleteCmd = &cobra.Command{
	Use:   "delete [file_id...]",
	Short: "Permanently delete items from recycle bin",
	Long:  "Permanently delete items from recycle bin. This action cannot be undone!",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		rb, ok := driver.(core.RecycleBin)
		if !ok {
			return fmt.Errorf("recycle bin is not supported by this driver")
		}

		err = rb.DeleteRecycle(cmd.Context(), args)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Permanently deleted %d item(s).\n", len(args))
		return nil
	},
}

func init() {
	recycleListCmd.Flags().IntP("page", "p", 1, "Page number")
	recycleListCmd.Flags().IntP("size", "s", 50, "Items per page")

	recycleCmd.AddCommand(recycleListCmd)
	recycleCmd.AddCommand(recycleRestoreCmd)
	recycleCmd.AddCommand(recycleDeleteCmd)
	rootCmd.AddCommand(recycleCmd)
}
