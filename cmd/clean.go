package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/trknhr/ai-docs/config"
	"github.com/trknhr/ai-docs/utils"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove AI docs worktree and branch",
	Long:  `Removes the AI docs worktree and optionally deletes the branch after confirmation.`,
	RunE:  runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {
	if !utils.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	docBranch := cfg.GetDocBranchName()

	if !force {
		fmt.Printf("This will remove the worktree at '%s' and the branch '%s'.\n", cfg.DocWorktreeDir, docBranch)
		fmt.Print("Are you sure? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Clean cancelled")
			return nil
		}
	}

	if dryRun {
		printWarning("Dry run mode - showing what would be done")
		fmt.Printf("Would remove worktree: %s\n", cfg.DocWorktreeDir)
		fmt.Printf("Would delete branch: %s\n", docBranch)
		return nil
	}

	for _, path := range cfg.AIAgentMemoryContextPath {
		if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
			printInfo("Removing symlink: %s", path)
			if err := os.Remove(path); err != nil {
				printWarning("Failed to remove symlink %s: %v", path, err)
			}
		}
	}

	if utils.PathExists(cfg.DocWorktreeDir) {
		printInfo("Removing worktree: %s", cfg.DocWorktreeDir)
		if err := utils.RunGit("", "worktree", "remove", "-f", cfg.DocWorktreeDir); err != nil {
			printWarning("Git worktree remove failed: %v", err)
			printInfo("Attempting manual removal")
			if err := os.RemoveAll(cfg.DocWorktreeDir); err != nil {
				return fmt.Errorf("failed to remove worktree directory: %w", err)
			}
		}
		printSuccess("Removed worktree")
	} else {
		printInfo("Worktree directory does not exist")
	}

	if utils.BranchExists(docBranch) {
		currentBranch, err := utils.GetCurrentBranch()
		if err == nil && currentBranch == docBranch {
			printInfo("Switching away from doc branch")
			if err := utils.RunGit("", "switch", cfg.MainBranchName); err != nil {
				return fmt.Errorf("failed to switch branch: %w", err)
			}
		}

		printInfo("Deleting branch: %s", docBranch)
		if err := utils.RunGit("", "branch", "-D", docBranch); err != nil {
			return fmt.Errorf("failed to delete branch: %w", err)
		}
		printSuccess("Deleted branch")

		printInfo("Deleting remote branch")
		if err := utils.RunGit("", "push", "origin", "--delete", docBranch); err != nil {
			printWarning("Failed to delete remote branch: %v", err)
		} else {
			printSuccess("Deleted remote branch")
		}
	} else {
		printInfo("Branch does not exist")
	}

	printSuccess("Clean completed successfully!")
	return nil
}
