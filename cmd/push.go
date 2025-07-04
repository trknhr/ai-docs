package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/trknhr/ai-docs/config"
	"github.com/trknhr/ai-docs/utils"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local AI docs to remote branch",
	Long:  `Copies local AI docs to the worktree, commits changes, and pushes to the remote repository.`,
	RunE:  runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	if !utils.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	printStep(1, 6, "Loading configuration")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	docBranch := cfg.GetDocBranchName()
	printInfo("Doc branch: %s", docBranch)
	printInfo("Worktree dir: %s", cfg.DocWorktreeDir)

	printStep(2, 6, "Validating worktree")
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

	printStep(3, 6, "Copying files to worktree")
	copiedCount := 0
	skippedCount := 0

	for _, path := range cfg.AIAgentMemoryContextPath {
		src := filepath.Join(".", path)
		dst := filepath.Join(cfg.DocWorktreeDir, path)

		if !utils.PathExists(src) {
			printInfo("Source path does not exist: %s (skipping)", src)
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

	printInfo("Files copied: %d, skipped: %d", copiedCount, skippedCount)

	printStep(4, 6, "Staging changes")
	if err := utils.RunGit(cfg.DocWorktreeDir, "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	if !utils.HasUncommittedChanges(cfg.DocWorktreeDir) {
		printInfo("No changes to commit")
		return nil
	}

	printStep(5, 6, "Creating commit")
	timestamp := time.Now().Format("2006-01-02_15:04:05")
	commitMsg := fmt.Sprintf("Update AI docs %s", timestamp)

	if err := utils.RunGit(cfg.DocWorktreeDir, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	printSuccess("Created commit: %s", commitMsg)

	printStep(6, 6, "Pushing to remote")
	if err := utils.PushWithRetry(cfg.DocWorktreeDir, docBranch, 3); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	printSuccess("Pushed changes to origin/%s", docBranch)

	return nil
}
