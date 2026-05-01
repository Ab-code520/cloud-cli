package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <source> <destination>",
	Short: "Synchronize directories (Local <-> Cloud)",
	Long: `Synchronize files between local directory and cloud drive.
Supports one-way sync (up/down) and deletion of extra files.

Examples:
  # Upload local folder to quark (only new/changed files)
  cloud-cli sync ./photos quark:/backup/photos

  # Preview sync (Dry Run)
  cloud-cli sync ./data quark:/backup --dry-run

  # Sync and delete extra files in destination
  cloud-cli sync ./important quark:/important --delete`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcInput := args[0]
		destInput := args[1]

		// Parse paths
		srcDriverName, srcPath, err := parseResourcePath(srcInput)
		if err != nil {
			return err
		}
		destDriverName, destPath, err := parseResourcePath(destInput)
		if err != nil {
			return err
		}

		// Initialize drivers
		// For local driver, if path is absolute, we set root to "/" so we can traverse fully.
		// If relative, we keep default (cwd) and pass relative path.
		srcRoot := ""
		if srcDriverName == "local" && filepath.IsAbs(srcPath) {
			srcRoot = "/"
		}
		srcDriver, err := initDriver(srcDriverName, srcRoot)
		if err != nil {
			return fmt.Errorf("init source driver: %w", err)
		}
		defer srcDriver.Close()

		destRoot := ""
		if destDriverName == "local" && filepath.IsAbs(destPath) {
			destRoot = "/"
		}
		destDriver, err := initDriver(destDriverName, destRoot)
		if err != nil {
			return fmt.Errorf("init dest driver: %w", err)
		}
		defer destDriver.Close()

		// Config
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		deleteExtra, _ := cmd.Flags().GetBool("delete")

		syncer := &core.Syncer{
			Source: srcDriver,
			Dest:   destDriver,
			Config: core.SyncConfig{
				DryRun:      dryRun,
				DeleteExtra: deleteExtra,
			},
			OnAction: func(action *core.SyncAction) {
				icon := " "
				switch action.Type {
				case "upload":
					icon = "⬆️"
				case "update":
					icon = "🔄"
				case "delete":
					icon = "🗑️"
				case "mkdir":
					icon = "📁"
				case "skip":
					icon = "⏭️"
				}
				
				msg := fmt.Sprintf("%s %s %s", icon, action.Type, action.Object.Name)
				if dryRun {
					msg += " [DRY RUN]"
				}
				fmt.Println(msg)
			},
		}

		fmt.Printf("🔄 Syncing %s -> %s ...\n", srcInput, destInput)
		if dryRun {
			fmt.Println("👀 This is a dry run, no files will be modified.")
		}

		if err := syncer.Sync(cmd.Context(), srcPath, destPath); err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		fmt.Println("✅ Sync completed.")
		return nil
	},
}

func init() {
	syncCmd.Flags().Bool("dry-run", false, "Preview changes without executing")
	syncCmd.Flags().Bool("delete", false, "Delete extra files in destination")
	rootCmd.AddCommand(syncCmd)
}
