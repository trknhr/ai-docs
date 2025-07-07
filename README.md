# AI Docs CLI

A Go-based CLI tool that provides a one-command workflow to isolate AI-generated "memory" files (GEMINI.md, CLAUDE.md etc) onto a dedicated Git branch with worktree and easy sync.

## Features

- Creates an orphan branch for AI memory files(GEMINI.md, CLAUDE.md, .cursor/rules/)
- Sets up a Git worktree for isolated file management  
- Automatically updates .gitignore
- Separate pull/push commands for flexible workflow
- Support for multiple AI agents (Cline, Claude, Gemini, Cursor)
- Configuration via YAML, JSON, or TOML

## Installation

### Using Homebrew

```bash
brew tap trknhr/homebrew-tap
brew install ai-docs
```

### Using install script

```bash
curl -sSL https://raw.githubusercontent.com/trknhr/ai-docs/main/install.sh | sh
```

### Using Go

```bash
go install github.com/trknhr/ai-docs@latest
```

### Build from source

```bash
git clone https://github.com/trknhr/ai-docs
cd ai-docs
go build -o ai-docs .
```

## Usage

### Initialize AI docs

```bash
ai-docs init [--config path/to/.ai-docs.config.yml] [--force] [--dry-run] [-v]
```

This command:
1. Reads configuration (defaults to `.ai-docs.config.yml`)
2. Creates an orphan branch (e.g., `@ai-docs/username`)
3. Copies AI memory files to the branch
4. Creates initial commit and pushes
5. Returns to main branch
6. Updates .gitignore
7. Adds a worktree at `.ai-docs`

> **⚠️ Important Notices**: 
> - If the config file doesn't exist, `ai-docs init` will create a sample `.ai-docs.config.yml` and exit. You'll need to review/edit the config file and run `ai-docs init` again.
> - If your AI memory files (like `CLAUDE.md`, `memory-bank/`, etc.) are not yet committed to any branch, they will be moved to the `@ai-docs/username` branch and will no longer exist in your main branch after running `init`. This is intentional to keep your main branch clean. You can use `ai-docs pull` to copy these files back to your main branch if needed.

### Push changes

```bash
ai-docs push [--config path/to/config.yml] [--dry-run] [-v]
```

Copies local AI docs to the worktree, commits and pushes ( to the `@ai-docs/username` branch ) changes to remote.

### Pull changes

```bash
ai-docs pull [--config path/to/config.yml] [--overwrite] [--dry-run] [-v]
```

Pulls latest changes from remote AI docs branch and copies them to your local project. Use `--overwrite` to replace existing local files.

### Clean up

```bash
ai-docs clean [--config path/to/config.yml] [--force] [--dry-run] [-v]
```

Removes the worktree and branch after confirmation.

## Configuration

Create `.ai-docs.config.yml` in your project root:

```yaml
userName: ""   # fallback when git config user.name is empty
mainBranchName: "main"

docBranchNameTemplate: "@ai-docs/{userName}"  # {userName} is replaced at runtime
docWorktreeDir: ".ai-docs"

aIAgentMemoryContextPath:
  Cline: "memory-bank"
  Claude: "CLAUDE.md"
  Gemini: "GEMINI.md"
  Cursor: ".cursor/rules"

ignorePatterns:
  - "/memory-bank/"
  - "CLAUDE.md"
  - "GEMINI.md"
  - "/.cursor/rules"
```

## Requirements

- Git 2.7.0+ (for worktree support)
- Go 1.24+ (for building from source)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright 2025 Teruo Kunihiro
