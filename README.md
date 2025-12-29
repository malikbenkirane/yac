# yac

Yet another git commit redactor in chief.

## AI-powered commit message generator

Implements a CLI tool that generates conventional commit messages using
Claude AI via Google Cloud Vertex AI. The tool analyzes git diffs and
repository context to produce structured, informative commit messages
following conventional commit format.

### Core capabilities:
- Integration with Vertex AI's Claude Sonnet 4.5 model for commit
  message generation
- Support for both git and jujutsu version control systems
- Configurable commit scope (api, auth, db, etc.) via flags or YAML/JSON
  configuration files
- Work-in-progress context tracking to document known issues, planned
  improvements, and implementation notes
- Git log context analysis to reference related commits
- Prompt template system for consistent AI interactions
- Debug mode that persists prompts and configuration for review

### Architecture:
- `cmd/` contains Cobra CLI commands for commit workflows
- `internal/agent/` handles AI prompt construction and templating
- `internal/commit/config/` manages YAML/JSON configuration with scope
  and WIP context unmarshaling
- `internal/commit/scope/` defines conventional commit scopes
- `internal/commit/wip/` categorizes work-in-progress notes (blockers,
  testing needs, technical debt, etc.)
- `internal/snake/` provides case conversion utilities

### The tool generates commit messages by:
1. Collecting git diff output and optional git log context
2. Building a structured prompt with diff, scope, and WIP notes
3. Posting to Vertex AI Claude API with OAuth token refresh
4. Extracting commit message from response via jq
5. Writing to `.commit-stash` for review before committing

Configuration can be prepared via `--prepare` flag or read from
`.prepare.yaml`/`.prepare.json` files, allowing users to specify commit
scope and document WIP context (known issues, planned improvements,
testing needs) that gets incorporated into the final commit message
body.

### Dependencies:
- spf13/cobra for CLI framework
- uber-go/zap for structured logging
- gopkg.in/yaml.v3 for YAML config parsing

