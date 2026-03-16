# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

ctxlog — a lightweight CLI tool for persistent, sharded context logging across AI agent sessions. Per-shard JSONL files with full CRUD and BSD `flock` for safe concurrent access. Zero external dependencies (Go standard library only).

## Build & Test

```bash
# Build
go build -o ctxlog .

# Run the binary
./ctxlog append -shard="test" -msg="hello world" -agent="claude"
./ctxlog read -shard="test" -lines=5
./ctxlog update -shard="test" -line=1 -msg="updated message"
./ctxlog delete -shard="test" -line=1
./ctxlog clear -shard="test"

# Install agent skill
./ctxlog install -type=claude
```

## Architecture

CLI binary (`main.go`) wraps the `memory` package. All data lives in `.ctxlog/` within the current working directory as per-shard JSONL files.

- `main.go` — CLI entry point. Parses subcommands and flags, dispatches to `memory.Store`.
- `memory/memory.go` — Core `Store` type. Manages `.ctxlog/` directory with per-shard JSONL files. Uses BSD `flock` for cross-process file locking (exclusive for writes, shared for reads). CRUD: `Append`, `ReadAll`, `ReadRecent`, `Update`, `Delete`, `Clear`.
- `skills/<agent>/SKILL.md` — Per-agent skill prompts, embedded into the binary via `go:embed`.
- `ctxlog install -type=<agent>` — Installs the skill for the given agent. Checks the agent's config dir exists first (e.g. `~/.claude`).

Entry format: `{"ts": <unix_seconds>, "agent": "<id>", "msg": "<text>"}`.

## Platform

Requires macOS or Linux (uses `syscall.Flock` for file locking).
