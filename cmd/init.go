package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

	printStep(1, 10, "Reading configuration")

	// Check if config file exists, create scaffolding if not
	if configPath == "" {
		configPath = "ai-docs.config.yml"
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

	printStep(2, 10, "Performing checks")
	if !utils.BranchExists(cfg.MainBranchName) {
		return fmt.Errorf("main branch '%s' does not exist", cfg.MainBranchName)
	}

	if utils.BranchExists(docBranch) && !force {
		return fmt.Errorf("doc branch '%s' already exists (use --force to override)", docBranch)
	}

	if utils.PathExists(cfg.DocWorktreeDir) && !force {
		return fmt.Errorf("worktree directory '%s' already exists (use --force to override)", cfg.DocWorktreeDir)
	}

	currentBranch, err := utils.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if dryRun {
		printWarning("Dry run mode - no changes will be made")
		return nil
	}

	printStep(3, 10, fmt.Sprintf("Creating docs branch: %s", docBranch))
	if utils.BranchExists(docBranch) && force {
		printInfo("Deleting existing branch: %s", docBranch)
		if err := utils.RunGit("", "branch", "-D", docBranch); err != nil {
			return fmt.Errorf("failed to delete existing branch: %w", err)
		}
	}

	// 🔐 Stash uncommitted changes before switching
	if err := utils.RunGit("", "stash", "-u"); err != nil {
		return fmt.Errorf("failed to stash before orphan switch: %w", err)
	}
	defer func() {
		_ = utils.RunGit("", "stash", "pop")
	}()

	if err := utils.RunGit("", "switch", "--orphan", docBranch); err != nil {
		return fmt.Errorf("failed to create orphan branch: %w", err)
	}

	// ✅ Validate that HEAD is now on the orphan branch
	currentBranchNow, err := utils.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to confirm current branch after switch: %w", err)
	}

	if currentBranchNow != docBranch {
		return fmt.Errorf("not on expected orphan branch '%s' (got '%s')", docBranch, currentBranchNow)
	}

	files, _ := filepath.Glob("*")
	for _, file := range files {
		if file != ".git" {
			if err := os.RemoveAll(file); err != nil {
				printWarning("Failed to remove %s: %v", file, err)
			}
		}
	}

	files, _ = filepath.Glob(".*")
	for _, file := range files {
		if file != "." && file != ".." && file != ".git" {
			os.RemoveAll(file)
		}
	}

	printStep(4, 10, "Copying AI directories")
	for agent, path := range cfg.AIAgentMemoryContextPath {
		printInfo("Processing %s: %s", agent, path)

		srcPath := filepath.Join("..", path)
		if utils.PathExists(srcPath) {
			if err := os.MkdirAll(path, 0755); err != nil {
				printWarning("Failed to create directory %s: %v", path, err)
				continue
			}

			if err := utils.CopyDir(srcPath, path); err != nil {
				printWarning("Failed to copy %s: %v", path, err)
			} else {
				printInfo("Copied %s", path)
			}
		} else {
			if err := os.MkdirAll(path, 0755); err != nil {
				printWarning("Failed to create directory %s: %v", path, err)
			}
		}
	}

	if cfg.DocDir != "" {
		if err := os.MkdirAll(cfg.DocDir, 0755); err != nil {
			printWarning("Failed to create doc directory %s: %v", cfg.DocDir, err)
		}
	}

	printStep(5, 10, "Creating initial commit")
	if err := utils.RunGit("", "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	if err := utils.RunGit("", "commit", "-m", "Initial AI docs commit", "--allow-empty"); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	if err := utils.PushWithRetry("", docBranch, 3); err != nil {
		printWarning("Failed to push branch: %v", err)
	} else {
		printSuccess("Pushed branch to origin")
	}

	printStep(6, 10, "Returning to main branch")
	if err := utils.RunGit("", "switch", cfg.MainBranchName); err != nil {
		if err := utils.RunGit("", "switch", currentBranch); err != nil {
			return fmt.Errorf("failed to return to branch: %w", err)
		}
	}

	printStep(7, 10, "Updating .gitignore")
	gitignorePath := ".gitignore"
	modified := false

	for _, pattern := range cfg.IgnorePatterns {
		if !utils.FileContains(gitignorePath, pattern) {
			if err := utils.AppendToFile(gitignorePath, []string{pattern}); err != nil {
				printWarning("Failed to add pattern to .gitignore: %v", err)
			} else {
				modified = true
				printInfo("Added to .gitignore: %s", pattern)
			}
		}
	}

	if !utils.FileContains(gitignorePath, cfg.DocWorktreeDir) {
		if err := utils.AppendToFile(gitignorePath, []string{cfg.DocWorktreeDir}); err != nil {
			printWarning("Failed to add worktree dir to .gitignore: %v", err)
		} else {
			modified = true
			printInfo("Added to .gitignore: %s", cfg.DocWorktreeDir)
		}
	}

	// if modified {
	// 	if err := utils.RunGit("", "add", ".gitignore"); err != nil {
	// 		printWarning("Failed to stage .gitignore: %v", err)
	// 	} else {
	// 		if err := utils.RunGit("", "commit", "-m", "Update .gitignore for AI docs"); err != nil {
	// 			printWarning("Failed to commit .gitignore: %v", err)
	// 		} else {
	// 			if err := utils.PushWithRetry("", cfg.MainBranchName, 3); err != nil {
	// 				printWarning("Failed to push .gitignore changes: %v", err)
	// 			}
	// 		}
	// 	}
	// }

	printStep(8, 10, "Adding worktree")
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

	printStep(9, 10, "Creating symlinks")
	for agent, path := range cfg.AIAgentMemoryContextPath {
		from := path
		to, err := filepath.Abs(filepath.Join(cfg.DocWorktreeDir, path))
		if err != nil {
			printWarning("Failed to get absolute path for %s: %v", path, err)
			continue
		}

		if err := utils.EnsureSymlink(from, to); err != nil {
			printWarning("Failed to create symlink for %s: %v", agent, err)
		} else {
			printSuccess("Created symlink: %s → %s", from, to)
		}
	}

	printStep(10, 10, "Initialization complete")
	printSuccess("AI docs initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  - Edit AI memory files in the symlinked directories")
	fmt.Println("  - Run 'ai-docs sync' to commit and push changes")

	return nil
}

func createScaffoldingConfig(path string) error {
	content := `userName: ""   # fallback when git config user.name is empty
mainBranchName: "main"

docBranchNameTemplate: "@doc/{userName}"  # {userName} ↔ runtime replace
docWorktreeDir: ".mem"

aIAgentMemoryContextPath:
  Cline: "memory-bank"
  Claude: "CLAUDE.md"
  Gemini: "GEMEMINI.md"
  Cursor: ".cursor/rules"

ignorePatterns:
  - "./memory-bank/"
  - "CLAUDE.md"
  - "GEMEMINI.md"
  - "./.cursor/rules/"
`
	return os.WriteFile(path, []byte(content), 0644)
}

// package utils

// import (
// 	"fmt"
// 	"os"
// )

// // EnsureSymlinkIfExists replaces existing file/dir/symlink at `from` with a symlink to `to`.
// // Does nothing if `from` does not exist.
// func EnsureSymlinkIfExists(from, to string) error {
// 	_, err := os.Lstat(from)
// 	if os.IsNotExist(err) {
// 		// Path does not exist → do nothing
// 		return nil
// 	}
// 	if err != nil {
// 		return fmt.Errorf("failed to stat %s: %w", from, err)
// 	}

// 	// Remove existing file/directory/symlink
// 	if err := os.RemoveAll(from); err != nil {
// 		return fmt.Errorf("failed to remove existing path at %s: %w", from, err)
// 	}

// 	// Create symlink
// 	if err := os.Symlink(to, from); err != nil {
// 		return fmt.Errorf("failed to create symlink from %s to %s: %w", from, to, err)
// 	}

// 	return nil
// }
