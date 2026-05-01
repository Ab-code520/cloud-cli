package cmd

import (
	"fmt"

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

		space, err := driver.Quota(cmd.Context())
		if err != nil {
			return err
		}

		totalGB := float64(space.Total) / 1024 / 1024 / 1024
		usedGB := float64(space.Used) / 1024 / 1024 / 1024
		freeGB := totalGB - usedGB
		usagePercent := float64(space.Used) / float64(space.Total) * 100

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

		return nil
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
