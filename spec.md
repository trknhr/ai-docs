# AI Docs CLI – Specification (v0.1)

> *Goal*: provide a one‑command workflow that isolates AI‑generated “memory” files onto a dedicated Git branch+worktree, with automatic symlinks and easy sync.
> *Target language*: **Go** (modules, Go 1.22+)

---

## 1  Overview

| Item             | Value                                                       |
| ---------------- | ----------------------------------------------------------- |
| **Main branch**  | `<mainBranchName>` (default `main`)                         |
| **Docs branch**  | `@doc/{userName}` (templated)                               |
| **Worktree dir** | `.mem` (configurable)                                       |
| **Config file**  | `ai-docs.config.yml` at repo root (override via `--config`) |
| **Init cmd**     | `ai-docs init`                                              |
| **Sync cmd**     | `ai-docs sync`                                              |

The binary reads a YAML/JSON/TOML config, creates an orphan branch that contains **only** the AI memory directories, commits & pushes it, appends ignore rules to the main branch, adds a worktree, makes symlinks back to the project root, then pushes everything. Future runs of `ai-docs sync` commit/push changes from the worktree.

---

## 2  Configuration (`ai-docs.config.yml`)

```yaml
userName: "defaultAsGitUsername"   # fallback when git config user.name is empty
mainBranchName: "main"

docBranchNameTemplate: "@doc/{userName}"  # {userName} ↔ runtime replace
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

*YAML is primary; JSON / TOML accepted via file extension.*

---

## 3  CLI Commands

### `ai-docs init`

| Phase                    | Operation                                                                                            |
| ------------------------ | ---------------------------------------------------------------------------------------------------- |
| ① **Read config**        | Parse file → fill defaults (see §2)                                                                  |
| ② **Checks**             | Ensure `mainBranch` exists; fail if `docBranch` or `docWorktreeDir` already exist (unless `--force`) |
| ③ **Create docs branch** | `git switch --orphan <docBranch>` → delete all files                                                 |
| ④ **Copy AI dirs**       | create/copy each `aIAgentMemoryContextPath` target                                                   |
| ⑤ **Initial commit**     | commit & push `origin <docBranch>`                                                                   |
| ⑥ **Return to main**     | `git switch <mainBranch>`                                                                            |
| ⑦ **Update .gitignore**  | append `ignorePatterns` if missing → commit & push                                                   |
| ⑧ **Add worktree**       | `git worktree add <docWorktreeDir> <docBranch>`                                                      |
| ⑨ **Symlink**            | for each memory dir: create symlink in project root → worktree path                                  |
| ⑩ **Done**               | print next‑steps hint (`ai-docs sync`)                                                               |

Options:

```
ai-docs init [--config <path>] [--dry-run] [-v|--verbose] [--force]
```

### `ai-docs sync`

```bash
git -C <docWorktreeDir> pull --quiet || true
git -C <docWorktreeDir> add -A
if ! git -C <docWorktreeDir> diff --cached --quiet; then
  git -C <docWorktreeDir> commit -m "sync ai docs $(date +%F_%T)"
  git -C <docWorktreeDir> push origin <docBranch>
fi
```

Options identical to `init` (without `--force`).

### `ai-docs clean`

*Optional helper* – removes worktree and branch after confirmation.

---

## 4  Error Handling

| Scenario                               | Behaviour                                          |
| -------------------------------------- | -------------------------------------------------- |
| Doc branch exists but worktree missing | Print hint, suggest `ai-docs sync`, exit 1         |
| Worktree exists but wrong branch       | Warn → exit 1 (`--force` overrides)                |
| Symlink creation fails                 | Detect OS, retry with junction (Windows) then warn |
| Push rejects                           | Retry ×3 with back‑off, else exit 1                |

---

## 5  Implementation Notes (Go)

### 5.1  Dependencies

* **YAML/TOML/JSON**: `github.com/go-yaml/yaml`, `github.com/pelletier/go-toml`
* **Exec Git**: `os/exec`, capture stdout/err
* **Path ops**: `os`, `os/user`, `filepath`
* **CLI flags**: `github.com/spf13/cobra` or std `flag`
* **Colorised output**: `github.com/fatih/color` (optional)

### 5.2  Key helpers

```go
func runGit(dir string, args ...string) error { /* exec git */ }
func fileContains(path, line string) bool     { /* … */ }
func ensureSymlink(from, to string) error      { /* cross‑platform */ }
```

### 5.3  Unit tests

* Use `t.TempDir()` with `git init` to simulate repos.
* Test idempotency: run `init` twice → expect no changes.
* Mock `exec.Command` for push failures.

---

## 6  Future Enhancements

1. **GitHub Actions scaffold** (`ai-docs ci-init`)
2. **Post‑commit hook auto‑sync** (`.git/hooks/post-commit`)
3. **Multiple‑user branches** via additional template vars `{date}` `{hostname}`.

---

© 2025 AI‑Docs Project – draft spec. Feel free to adapt for your team’s workflow.
