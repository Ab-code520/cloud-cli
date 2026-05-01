package cmd

import (
	"fmt"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for files and folders in cloud drive",
	Long: `Search for files and folders in your cloud drive.

Examples:
  cloud-cli search "report.pdf"         # Search by filename
  cloud-cli search "photos" --dir abc   # Search in specific directory
  cloud-cli search "*.mp4"              # Use wildcard pattern`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		s, ok := driver.(core.Searchable)
		if !ok {
			return fmt.Errorf("search is not supported by this driver")
		}

		dirID, _ := cmd.Flags().GetString("dir")
		page, _ := cmd.Flags().GetInt("page")
		size, _ := cmd.Flags().GetInt("size")

		results, err := s.Search(cmd.Context(), query, dirID, page, size)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Printf("🔍 No results found for '%s'\n", query)
			return nil
		}

		fmt.Printf("🔍 Found %d results for '%s':\n\n", len(results), query)
		for _, obj := range results {
			typeIcon := "📄"
			if obj.IsDir {
				typeIcon = "📁"
			}
			sizeStr := ""
			if !obj.IsDir {
				sizeStr = fmt.Sprintf(" (%s)", formatSize(obj.Size))
			}
			fmt.Printf("%s %s%s\n", typeIcon, obj.Name, sizeStr)
			fmt.Printf("   🆔 %s | 📅 %s\n\n", obj.ID, obj.ModTime.Format("2006-01-02"))
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().StringP("dir", "d", "", "Search in specific directory ID")
	searchCmd.Flags().IntP("page", "p", 1, "Page number")
	searchCmd.Flags().IntP("size", "s", 50, "Results per page")
	rootCmd.AddCommand(searchCmd)
}
