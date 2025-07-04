package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/trknhr/ai-docs/config"
	"github.com/trknhr/ai-docs/utils"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AI docs branch and worktree",
	Long:  `Creates an orphan branch for AI memory files, sets up worktree, and creates symlinks.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&force, "force", false, "force initialization even if branch/worktree exists")
}

func runInit(cmd *cobra.Command, args []string) error {
	if !utils.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	printStep(1, 9, "Reading configuration")

	// Check if config file exists, create scaffolding if not
	if configPath == "" {
		configPath = ".ai-docs.config.yml"
	}

	if !utils.PathExists(configPath) {
		printWarning("Config file not found at: %s", configPath)
		if err := createScaffoldingConfig(configPath); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		printSuccess("Created sample config file: %s", configPath)
		fmt.Println("\nPlease review and edit the configuration file, then run 'ai-docs init' again.")
		return nil
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	printInfo("Loaded config from: %s", configPath)

	docBranch := cfg.GetDocBranchName()
	printInfo("Doc branch: %s", docBranch)
	printInfo("Worktree dir: %s", cfg.DocWorktreeDir)

	printStep(2, 9, "Performing checks")
	if !utils.BranchExists(cfg.MainBranchName) {
		return fmt.Errorf("main branch '%s' does not exist", cfg.MainBranchName)
	}

	if utils.BranchExists(docBranch) && !force {
		return fmt.Errorf("doc branch '%s' already exists (use --force to override)", docBranch)
	}

	if utils.PathExists(cfg.DocWorktreeDir) {
		if force {
			cmd := exec.Command("git", "worktree", "remove", "--force", cfg.DocWorktreeDir)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to remove worktree %s: %w", cfg.DocWorktreeDir, err)
			}
		} else {
			return fmt.Errorf("worktree directory '%s' already exists (use --force to override)", cfg.DocWorktreeDir)
		}
	}

	currentBranch, err := utils.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if dryRun {
		printWarning("Dry run mode - no changes will be made")
		return nil
	}

	printStep(3, 9, fmt.Sprintf("Creating docs branch: %s", docBranch))
	if utils.BranchExists(docBranch) && force {
		printInfo("Deleting existing branch: %s", docBranch)
		if err := utils.RunGit("", "branch", "-D", docBranch); err != nil {
			return fmt.Errorf("failed to delete existing branch: %w", err)
		}
	}

	if err := utils.RunGit("", "checkout", "--orphan", docBranch); err != nil {
		return fmt.Errorf("failed to create orphan branch: %w", err)
	}
	if err := utils.RunGit("", "reset"); err != nil {
		return fmt.Errorf("failed to rest orphan branch: %w", err)
	}

	// ✅ Validate that HEAD is now on the orphan branch
	printStep(4, 9, "Validating branch switch")
	currentBranchNow, err := utils.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to confirm current branch after switch: %w", err)
	}

	if currentBranchNow != docBranch {
		return fmt.Errorf("not on expected orphan branch '%s' (got '%s')", docBranch, currentBranchNow)
	}
	printSuccess("Successfully switched to orphan branch: %s", docBranch)

	printStep(5, 9, "Creating initial commit")
	for _, path := range cfg.AIAgentMemoryContextPath {
		if utils.PathExists(path) {
			if err := utils.RunGit("", "add", "-f", path); err != nil {
				printWarning("Failed to stage %s: %v", path, err)
			} else {
				printInfo("Staged: %s", path)
			}
		} else {
			printInfo("Skipped (not found): %s", path)
		}
	}

	if err := utils.RunGit("", "commit", "-m", "Initial AI docs commit", "--allow-empty"); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	if err := utils.PushWithRetry("", docBranch, 3); err != nil {
		printWarning("Failed to push branch: %v", err)
	} else {
		printSuccess("Pushed branch to origin")
	}

	printStep(6, 9, "Returning to main branch")
	if err := utils.RunGit("", "switch", "-f", cfg.MainBranchName); err != nil {
		if err := utils.RunGit("", "switch", currentBranch); err != nil {
			return fmt.Errorf("failed to return to branch: %w", err)
		}
	}

	printStep(7, 9, "Updating .gitignore")
	gitignorePath := ".gitignore"

	for _, pattern := range cfg.IgnorePatterns {
		if !utils.FileContains(gitignorePath, pattern) {
			if err := utils.AppendToFile(gitignorePath, []string{pattern}); err != nil {
				printWarning("Failed to add pattern to .gitignore: %v", err)
			} else {
				printInfo("Added to .gitignore: %s", pattern)
			}
		}
	}

	if !utils.FileContains(gitignorePath, cfg.DocWorktreeDir) {
		if err := utils.AppendToFile(gitignorePath, []string{cfg.DocWorktreeDir}); err != nil {
			printWarning("Failed to add worktree dir to .gitignore: %v", err)
		} else {
			printInfo("Added to .gitignore: %s", cfg.DocWorktreeDir)
		}
	}

	printStep(8, 9, "Adding worktree")
	if utils.PathExists(cfg.DocWorktreeDir) && force {
		printInfo("Removing existing worktree")
		if err := utils.RunGit("", "worktree", "remove", "-f", cfg.DocWorktreeDir); err != nil {
			os.RemoveAll(cfg.DocWorktreeDir)
		}
	}

	if err := utils.RunGit("", "worktree", "add", cfg.DocWorktreeDir, docBranch); err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}
	printSuccess("Added worktree at %s", cfg.DocWorktreeDir)

	printStep(9, 9, "Initialization complete")
	printSuccess("AI docs initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  - Edit AI memory files in the symlinked directories")
	fmt.Println("  - Run 'ai-docs push' to commit and push changes")
	fmt.Println("  - Run 'ai-docs pull' to get latest changes from remote")

	return nil
}

func createScaffoldingConfig(path string) error {
	content := `userName: ""   # fallback git config user.name or whoami when userName is embpty 
mainBranchName: "main"

docBranchNameTemplate: "@ai-docs/{userName}"  # {userName} ↔ runtime replace
docWorktreeDir: ".ai-docs"

aIAgentMemoryContextPath:
  Cline: "memory-bank"
  Claude: "CLAUDE.md"
  Gemini: "GEMINI.md"
  Cursor: ".cursor/rules"

ignorePatterns:
  - "memory-bank/"
  - "CLAUDE.md"
  - "GEMINI.md"
  - ".cursor/rules/"
`
	return os.WriteFile(path, []byte(content), 0644)
}
