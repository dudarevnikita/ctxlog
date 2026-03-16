# ctxlog

Lightweight CLI tool for persistent, sharded context logging across AI agent sessions. Per-shard JSONL files with BSD `flock` for safe concurrent access. Zero external dependencies — only Go standard library.

## Requirements

- Go 1.22+
- macOS or Linux (uses BSD `flock` for cross-process file locking)

## Install

```bash
brew tap dudarevnikita/tools
brew install ctxlog
```

## Build from source

```bash
go build -o ctxlog .
```

## Usage

### Append an entry

```bash
ctxlog append -shard="auth" -msg="Fixed DB connection pooling bug" -agent="claude"
```

`-agent` is optional. `-shard` and `-msg` are required.

Status is printed to stderr:

```
ok: appended to .ctxlog/auth.jsonl
```

### Read recent entries

```bash
ctxlog read -shard="auth" -lines=5
```

`-lines` defaults to 10. Output goes to stdout with numbered lines:

```
1: {"ts":1773685005,"agent":"claude","msg":"Fixed DB connection pooling bug"}
```

Returns empty output if the shard doesn't exist yet.

### Update an entry

```bash
ctxlog update -shard="auth" -line=2 -msg="Updated: connection pool size set to 25"
```

Updates the message and refreshes the timestamp at the given 1-based line number.

### Delete an entry

```bash
ctxlog delete -shard="auth" -line=3
```

Removes the entry at the given line number. Remaining lines are renumbered.

### Clear a shard

```bash
ctxlog clear -shard="auth"
```

Deletes the entire shard file.

### Install Claude Code skill

```bash
ctxlog install
```

Writes `SKILL.md` to `~/.claude/skills/ctxlog/` so Claude Code automatically knows how to use the tool across all projects.

## File structure on disk

```
<cwd>/
└── .ctxlog/
    ├── auth.jsonl
    └── tasks/
        └── task_123.jsonl
```

Each `.jsonl` file contains one JSON object per line:

```json
{"ts":1773685005,"agent":"claude","msg":"Fixed DB connection pooling bug"}
```

## Flags reference

| Command | Flag | Required | Default | Description |
|---------|------|----------|---------|-------------|
| `append` | `-shard` | yes | — | Shard name (supports `/` for nesting) |
| `append` | `-msg` | yes | — | Message to log |
| `append` | `-agent` | no | `""` | Agent identifier |
| `read` | `-shard` | yes | — | Shard name |
| `read` | `-lines` | no | `10` | Number of recent entries to return |
| `update` | `-shard` | yes | — | Shard name |
| `update` | `-line` | yes | — | 1-based line number to update |
| `update` | `-msg` | yes | — | New message text |
| `delete` | `-shard` | yes | — | Shard name |
| `delete` | `-line` | yes | — | 1-based line number to delete |
| `clear` | `-shard` | yes | — | Shard name to remove |
| `install` | — | — | — | No flags |

## Concurrency

Safe for concurrent use across multiple processes:

- BSD `flock` on each shard file — exclusive for writes, shared for reads
- `O_APPEND` mode for kernel-level write atomicity

## Project structure

```
├── go.mod
├── main.go           # CLI entry point: subcommands, flags, help, install
├── memory/
│   └── memory.go     # Store: Append, ReadAll, ReadRecent, Update, Delete, Clear
└── README.md
```
