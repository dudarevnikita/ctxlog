# ctxlog

Lightweight CLI tool for persistent, sharded context logging across AI agent sessions. Append-only JSONL files with BSD `flock` for safe concurrent access. Zero external dependencies — only Go standard library.

## Requirements

- Go 1.22+
- macOS or Linux (uses BSD `flock` for cross-process file locking)

## Build

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

`-lines` defaults to 10. Output goes to stdout as JSON:

```json
[
  {
    "ts": 1773685005,
    "agent": "claude",
    "msg": "Fixed DB connection pooling bug"
  }
]
```

Returns `[]` if the shard doesn't exist yet.

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
| `install` | — | — | — | No flags |

## Concurrency

Safe for concurrent use across multiple processes:

- BSD `flock` on each shard file — exclusive for writes, shared for reads
- `O_APPEND` mode for kernel-level write atomicity

## Project structure

```
├── go.mod
├── main.go           # CLI entry point: subcommands, flags, install
├── memory/
│   └── memory.go     # Store, Append, ReadRecent, flock
└── README.md
```
