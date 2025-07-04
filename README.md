# AI Docs CLI

A Go-based CLI tool that provides a one-command workflow to isolate AI-generated "memory" files onto a dedicated Git branch with worktree, automatic symlinks, and easy sync.

## Features

- Creates an orphan branch for AI memory files
- Sets up a Git worktree for isolated file management  
- Creates symlinks from project root to worktree
- Automatically updates .gitignore
- Separate pull/push commands for flexible workflow
- Support for multiple AI agents (Cline, Claude, Gemini, Cursor)
- Configuration via YAML, JSON, or TOML

## Installation

```bash
go install github.com/trknhr/ai-docs@latest
```

Or build from source:

```bash
git clone https://github.com/trknhr/ai-docs
cd ai-docs
go build -o ai-docs .
```

## Usage

### Initialize AI docs

```bash
ai-docs init [--config path/to/config.yml] [--force] [--dry-run] [-v]
```

This command:
1. Reads configuration
2. Creates an orphan branch (e.g., `@doc/username`)
3. Copies AI memory directories to the branch
4. Creates initial commit and pushes
5. Returns to main branch
6. Updates .gitignore
7. Adds a worktree at `.mem`
8. Creates symlinks from project root to worktree

### Push changes

```bash
ai-docs push [--config path/to/config.yml] [--dry-run] [-v]
```

Copies local AI docs to the worktree, commits and pushes changes to remote.

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

Create `ai-docs.config.yml` in your project root:

```yaml
userName: "defaultAsGitUsername"   # fallback when git config user.name is empty
mainBranchName: "main"

docBranchNameTemplate: "@doc/{userName}"  # {userName} is replaced at runtime
docWorktreeDir: ".mem"

aIAgentMemoryContextPath:
  Cline: "memory-bank"
  Claude: ".ai-memory"
  Gemini: ".gemini/context"
  Cursor: ".cursor/rules"

ignorePatterns:
  - "/memory-bank/"
  - "/.ai-memory/"
  - "/.gemini/context/"
  - "/.cursor/rules/"

docDir: "docs/ai"           # optional – where to add extra docs
```

## Requirements

- Git 2.7.0+ (for worktree support)
- Go 1.24+ (for building from source)

## License

See spec.md for more details.