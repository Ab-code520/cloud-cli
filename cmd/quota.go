package cmd

import (
	"fmt"

	"github.com/Ab-code520/cloud-cli/core"
	"github.com/spf13/cobra"
)

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Show cloud drive storage usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		driver, err := getDriver()
		if err != nil {
			return err
		}
		defer driver.Close()

		qp, ok := driver.(core.QuotaProvider)
		if !ok {
			return fmt.Errorf("storage quota is not supported by this driver")
		}

		total, used, err := qp.Quota(cmd.Context())
		if err != nil {
			return err
		}

		data := map[string]interface{}{
			"total_bytes": total,
			"used_bytes":  used,
			"free_bytes":  total - used,
			"total_gb":    float64(total) / 1024 / 1024 / 1024,
			"used_gb":     float64(used) / 1024 / 1024 / 1024,
			"free_gb":     float64(total-used) / 1024 / 1024 / 1024,
			"usage_pct":   float64(used) / float64(total) * 100,
		}

		return outputTableOrJSON(data, func() {
			totalGB := float64(total) / 1024 / 1024 / 1024
			usedGB := float64(used) / 1024 / 1024 / 1024
			freeGB := totalGB - usedGB
			usagePercent := float64(used) / float64(total) * 100

			fmt.Println("💾 Cloud Drive Storage Usage")
			fmt.Println("─────────────────────────────")
			fmt.Printf("📊 Total:   %.2f GB\n", totalGB)
			fmt.Printf("📦 Used:    %.2f GB (%.1f%%)\n", usedGB, usagePercent)
			fmt.Printf("🆓 Free:    %.2f GB\n", freeGB)
			fmt.Println("─────────────────────────────")

			// Visual bar
			barLen := 30
			filled := int(usagePercent / 100 * float64(barLen))
			if filled > barLen {
				filled = barLen
			}
			fmt.Printf("[%s%s] %.1f%%\n",
				repeatChar("█", filled),
				repeatChar("░", barLen-filled),
				usagePercent)
		})
	},
}

func repeatChar(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func init() {
	rootCmd.AddCommand(quotaCmd)
}
