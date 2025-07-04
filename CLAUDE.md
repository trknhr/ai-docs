# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build & Development
```bash
make build          # Build the ai-docs binary
make install        # Install to GOPATH/bin
make run ARGS="init -v"  # Build and run with arguments
make fmt            # Format all Go code
make lint           # Run golangci-lint (must be installed)
make deps           # Download and tidy dependencies
make deps-update    # Update all dependencies to latest versions
```

### Testing Commands
```bash
make test           # Run tests (note: no tests exist yet)
make test-coverage  # Run tests with coverage report
make init           # Quick test: run 'ai-docs init -v'
make pull           # Quick test: run 'ai-docs pull -v'
make push           # Quick test: run 'ai-docs push -v'
make sync           # Deprecated: run 'ai-docs sync -v'
```

### Clean Up
```bash
make clean          # Remove build artifacts
```

## Architecture Overview

This is a Go CLI tool that manages AI-generated memory files by isolating them on a dedicated Git branch with worktree support. It solves the problem of AI assistants' context files cluttering the main repository.

### Core Workflow

1. **Orphan Branch Strategy**: Creates a separate branch (`@ai-doc/username`) with no shared history from main, keeping AI files completely isolated
2. **Git Worktree**: Uses Git's worktree feature to maintain AI files in `.ai-docs` directory while keeping main branch clean
3. **Pull/Push Workflow**: Separate commands for pulling remote changes and pushing local updates, providing flexible synchronization

### Key Components

- **cmd/**: Cobra-based CLI commands (init, sync, clean)
  - Each command follows a step-by-step process with colored output
  - All commands support `--dry-run` for testing and `-v` for verbose output
  
- **config/**: Configuration system supporting YAML/JSON/TOML
  - Auto-detects username from git config or whoami
  - Template-based branch naming with `{userName}` substitution
  
- **utils/**: Core utilities for Git and file operations
  - `RunGit()`: Executes git commands with proper error handling
  - `PushWithRetry()`: Network-resilient push with exponential backoff
  - Cross-platform symlink support (Unix symlinks, Windows junctions)

### Important Implementation Details

1. **Git Operations**: All Git commands run via `os/exec` with stderr capture for error messages
2. **Branch Validation**: Always validates current branch state before and after operations
3. **Force Flags**: Both init and clean support `--force` for overriding existing states
4. **Config Scaffolding**: Creates example config on first run, then exits for user review

### Common Pitfalls

- The tool modifies `.gitignore` on the main branch - these changes need to be committed separately
- Symlink creation (step 9 in init) is not fully implemented yet
- No test files exist despite Makefile test targets
- Worktree directory must not exist before init (unless using --force)

### Development Notes

When modifying this codebase:
- Always test with `--dry-run` first
- Use `-v` flag to see actual Git commands being executed
- The orphan branch workflow is critical - never merge it with main
- Config file changes require re-running init to take effect
- Print functions (printStep, printInfo, etc.) in cmd/root.go handle colored output