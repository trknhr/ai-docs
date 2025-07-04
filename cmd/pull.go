package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/trknhr/ai-docs/config"
	"github.com/trknhr/ai-docs/utils"
)

var (
	overwrite bool
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull AI docs from remote branch to local",
	Long:  `Pulls latest changes from the remote AI docs branch and copies them to your local project.`,
	RunE:  runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite local files without warning")
}

func runPull(cmd *cobra.Command, args []string) error {
	if !utils.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	printStep(1, 5, "Loading configuration")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	docBranch := cfg.GetDocBranchName()
	printInfo("Doc branch: %s", docBranch)
	printInfo("Worktree dir: %s", cfg.DocWorktreeDir)

	printStep(2, 5, "Validating worktree")
	if !utils.PathExists(cfg.DocWorktreeDir) {
		return fmt.Errorf("worktree directory '%s' does not exist - run 'ai-docs init' first", cfg.DocWorktreeDir)
	}

	if !utils.BranchExists(docBranch) {
		return fmt.Errorf("doc branch '%s' does not exist - run 'ai-docs init' first", docBranch)
	}

	if dryRun {
		printWarning("Dry run mode - no changes will be made")
		return nil
	}

	printStep(3, 5, "Pulling from remote")
	printInfo("Pulling latest changes from origin/%s", docBranch)

	if err := utils.RunGit(cfg.DocWorktreeDir, "pull", "--quiet"); err != nil {
		printWarning("Pull failed (may be normal for new branches): %v", err)
	} else {
		printSuccess("Successfully pulled latest changes")
	}

	printStep(4, 5, "Copying files to local")
	copiedCount := 0
	skippedCount := 0

	for _, path := range cfg.AIAgentMemoryContextPath {
		src := filepath.Join(cfg.DocWorktreeDir, path)
		dst := filepath.Join(".", path)

		if !utils.PathExists(src) {
			printInfo("Remote file does not exist: %s (skipping)", path)
			skippedCount++
			continue
		}

		// Check if local file exists and warn user
		if utils.PathExists(dst) && !overwrite {
			printWarning("Local file exists: %s (use --overwrite to replace)", dst)
			skippedCount++
			continue
		}

		if err := utils.CopyPath(src, dst); err != nil {
			printWarning("Failed to copy %s → %s: %v", src, dst, err)
			skippedCount++
		} else {
			printSuccess("Copied: %s", path)
			copiedCount++
		}
	}

	printStep(5, 5, "Pull complete")
	printInfo("Files copied: %d, skipped: %d", copiedCount, skippedCount)

	if skippedCount > 0 && !overwrite {
		fmt.Println("\nUse --overwrite flag to replace existing local files")
	}

	return nil
}
