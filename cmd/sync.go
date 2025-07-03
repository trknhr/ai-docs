package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/trknhr/ai-docs/config"
	"github.com/trknhr/ai-docs/utils"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync AI docs to the remote branch",
	Long:  `Commits and pushes changes from the AI docs worktree to the remote repository.`,
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	if !utils.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	docBranch := cfg.GetDocBranchName()
	
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

	printInfo("Syncing worktree: %s", cfg.DocWorktreeDir)
	
	if err := utils.RunGit(cfg.DocWorktreeDir, "pull", "--quiet"); err != nil {
		printWarning("Pull failed (may be normal for new branches): %v", err)
	}

	if err := utils.RunGit(cfg.DocWorktreeDir, "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	if !utils.HasUncommittedChanges(cfg.DocWorktreeDir) {
		printInfo("No changes to commit")
		return nil
	}

	timestamp := time.Now().Format("2006-01-02_15:04:05")
	commitMsg := fmt.Sprintf("sync ai docs %s", timestamp)
	
	if err := utils.RunGit(cfg.DocWorktreeDir, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	printSuccess("Created commit: %s", commitMsg)

	if err := utils.PushWithRetry(cfg.DocWorktreeDir, docBranch, 3); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	printSuccess("Pushed changes to origin/%s", docBranch)

	return nil
}