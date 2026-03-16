// ctxlog — a lightweight CLI tool for persistent, sharded context logging.
//
// Stores append-only JSONL entries in .ctxlog/<shard>.jsonl within the
// current working directory. Uses BSD flock for cross-process safety.
//
// Usage:
//
//	ctxlog append -shard=<name> -msg=<text> [-agent=<id>]
//	ctxlog read   -shard=<name> [-lines=10]
//	ctxlog install
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ctxlog/memory"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("usage: ctxlog <append|read|install> [flags]")
	}

	switch os.Args[1] {
	case "append":
		cmdAppend(os.Args[2:])
	case "read":
		cmdRead(os.Args[2:])
	case "install":
		cmdInstall()
	default:
		fatalf("unknown command: %q (want append, read, or install)", os.Args[1])
	}
}

func cmdAppend(args []string) {
	fs := flag.NewFlagSet("append", flag.ExitOnError)
	shard := fs.String("shard", "", "shard name (required)")
	msg := fs.String("msg", "", "message to log (required)")
	agent := fs.String("agent", "", "agent identifier (optional)")
	fs.Parse(args)

	if *shard == "" || *msg == "" {
		fatalf("append: -shard and -msg are required")
	}

	store := storeFromCwd()

	err := store.Append(*shard, memory.Entry{
		Agent: *agent,
		Msg:   *msg,
	})
	if err != nil {
		fatalf("append: %v", err)
	}

	fmt.Fprintf(os.Stderr, "ok: appended to .ctxlog/%s.jsonl\n", *shard)
}

func cmdRead(args []string) {
	fs := flag.NewFlagSet("read", flag.ExitOnError)
	shard := fs.String("shard", "", "shard name (required)")
	lines := fs.Int("lines", 10, "number of recent lines to return")
	fs.Parse(args)

	if *shard == "" {
		fatalf("read: -shard is required")
	}

	store := storeFromCwd()

	entries, err := store.ReadRecent(*shard, *lines)
	if err != nil {
		fatalf("read: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(entries)
}

const skillContent = `## Metadata
name: Context Logger (ctxlog)
description: Apply long-term memory and cross-session state persistence using the ctxlog CLI tool.

## Overview
This Skill provides Claude with a persistent memory system. Because Claude loses context between sessions, ` + "`ctxlog`" + ` acts as a sharded, append-only local database. When working on complex tasks, Claude must use this tool to log progress, save bug workarounds, or read previous context.

## How to Use (Execution)
Execute the global CLI binary directly in the terminal shell:
- To write memory: ` + "`ctxlog append -shard=\"<task_id>\" -msg=\"<your_message>\"`" + `
- To read memory: ` + "`ctxlog read -shard=\"<task_id>\" -lines=10`" + `

## When to Apply
Apply this skill strictly when:
- You have completed a significant logical block of a task.
- You have discovered a bug workaround that you might need to remember later.
- You are starting a new session and need to recall what was done previously on a specific ` + "`<task_id>`" + `.

## Strict Rules
- NEVER invent or create your own markdown memory files (like ` + "`memory.md`" + ` or ` + "`notes.txt`" + `).
- ALWAYS rely exclusively on the ` + "`ctxlog`" + ` CLI tool for state persistence.
- Keep the ` + "`-msg`" + ` payload concise and factual.
`

func cmdInstall() {
	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("install: cannot determine home directory: %v", err)
	}

	dir := filepath.Join(home, ".claude", "skills", "ctxlog")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatalf("install: mkdir: %v", err)
	}

	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(skillContent), 0o644); err != nil {
		fatalf("install: write: %v", err)
	}

	fmt.Printf("installed skill to %s\n", path)
}

func storeFromCwd() *memory.Store {
	cwd, err := os.Getwd()
	if err != nil {
		fatalf("getwd: %v", err)
	}
	store := memory.NewStore(cwd)
	if err := store.Init(); err != nil {
		fatalf("init: %v", err)
	}
	return store
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
