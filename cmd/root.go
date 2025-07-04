package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	configPath string
	dryRun     bool
	verbose    bool
	force      bool
)

var rootCmd = &cobra.Command{
	Use:   "ai-docs",
	Short: "AI documentation management tool",
	Long: `AI Docs CLI provides a one-command workflow that isolates AI-generated "memory" files 
onto a dedicated Git branch+worktree, with automatic symlinks and easy sync.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (default: ai-docs.config.yml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func printInfo(format string, args ...interface{}) {
	if verbose {
		color.Blue(format, args...)
	}
}

func printSuccess(format string, args ...interface{}) {
	color.Green("✓ "+format, args...)
}

func printWarning(format string, args ...interface{}) {
	color.Yellow("⚠ "+format, args...)
}

func printStep(step int, total int, description string) {
	fmt.Printf("[%d/%d] %s\n", step, total, description)
}
