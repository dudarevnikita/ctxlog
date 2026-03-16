// ctxlog — a lightweight CLI tool for persistent, sharded context logging.
//
// Stores append-only JSONL entries in .ctxlog/<shard>.jsonl within the
// current working directory. Uses BSD flock for cross-process safety.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ctxlog/memory"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `ctxlog — persistent, sharded context logging for AI agent sessions

Usage:
  ctxlog <command> [flags]

Commands:
  append    Append a new entry to a shard
  read      Read recent entries from a shard
  update    Update an existing entry by line number
  delete    Delete an entry by line number
  clear     Remove an entire shard file
  install   Install the Claude Code skill to ~/.claude/skills/

Flags by command:

  append  -shard=<name>  -msg=<text>  [-agent=<id>]
  read    -shard=<name>  [-lines=10]
  update  -shard=<name>  -line=<num>  -msg=<new_text>
  delete  -shard=<name>  -line=<num>
  clear   -shard=<name>

Examples:
  ctxlog append -shard="auth-refactor" -msg="switched to JWT middleware" -agent="claude"
  ctxlog read   -shard="auth-refactor" -lines=5
  ctxlog update -shard="auth-refactor" -line=2 -msg="JWT middleware validated in staging"
  ctxlog delete -shard="auth-refactor" -line=3
  ctxlog clear  -shard="auth-refactor"
  ctxlog install
`)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" || os.Args[1] == "help" {
		printUsage()
		if len(os.Args) < 2 {
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "append":
		cmdAppend(os.Args[2:])
	case "read":
		cmdRead(os.Args[2:])
	case "update":
		cmdUpdate(os.Args[2:])
	case "delete":
		cmdDelete(os.Args[2:])
	case "clear":
		cmdClear(os.Args[2:])
	case "install":
		cmdInstall()
	default:
		fatalf("unknown command: %q\nRun 'ctxlog --help' for usage.", os.Args[1])
	}
}

func cmdAppend(args []string) {
	shard, msg, _, agent := parseFlags("append", args, true, true, false, true)

	store := storeFromCwd()
	err := store.Append(shard, memory.Entry{
		Agent: agent,
		Msg:   msg,
	})
	if err != nil {
		fatalf("append: %v", err)
	}

	fmt.Fprintf(os.Stderr, "ok: appended to .ctxlog/%s.jsonl\n", shard)
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

	for i, e := range entries {
		data, _ := json.Marshal(e)
		fmt.Printf("%d: %s\n", i+1, data)
	}
}

func cmdUpdate(args []string) {
	shard, msg, line, _ := parseFlags("update", args, true, true, true, false)

	store := storeFromCwd()
	if err := store.Update(shard, line, msg); err != nil {
		fatalf("update: %v", err)
	}

	fmt.Fprintf(os.Stderr, "ok: updated line %d in .ctxlog/%s.jsonl\n", line, shard)
}

func cmdDelete(args []string) {
	shard, _, line, _ := parseFlags("delete", args, true, false, true, false)

	store := storeFromCwd()
	if err := store.Delete(shard, line); err != nil {
		fatalf("delete: %v", err)
	}

	fmt.Fprintf(os.Stderr, "ok: deleted line %d from .ctxlog/%s.jsonl\n", line, shard)
}

func cmdClear(args []string) {
	shard, _, _, _ := parseFlags("clear", args, true, false, false, false)

	store := storeFromCwd()
	if err := store.Clear(shard); err != nil {
		fatalf("clear: %v", err)
	}

	fmt.Fprintf(os.Stderr, "ok: cleared .ctxlog/%s.jsonl\n", shard)
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
- To update a specific line: ` + "`ctxlog update -shard=\"<task_id>\" -line=<num> -msg=\"<new_text>\"`" + `
- To delete a specific line: ` + "`ctxlog delete -shard=\"<task_id>\" -line=<num>`" + `
- To wipe a shard: ` + "`ctxlog clear -shard=\"<task_id>\"`" + `

## When to Apply
Apply this skill strictly when:
- You have completed a significant logical block of a task.
- You have discovered a bug workaround that you might need to remember later.
- You are starting a new session and need to recall what was done previously on a specific ` + "`<task_id>`" + `.

## Strict Rules
- NEVER invent or create your own markdown memory files (like ` + "`memory.md`" + ` or ` + "`notes.txt`" + `).
- ALWAYS rely exclusively on the ` + "`ctxlog`" + ` CLI tool for state persistence.
- Keep the ` + "`-msg`" + ` payload concise and factual.
- If memory contains outdated or resolved issues, use ` + "`update`" + ` or ` + "`delete`" + ` to keep the context clean and save tokens.
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

func parseFlags(cmd string, args []string, needShard, needMsg, needLine, needAgent bool) (shard, msg string, line int, agent string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	shardP := fs.String("shard", "", "")
	msgP := fs.String("msg", "", "")
	lineP := fs.Int("line", 0, "")
	agentP := fs.String("agent", "", "")
	fs.Parse(args)

	if needShard && *shardP == "" {
		fatalf("%s: -shard is required", cmd)
	}
	if needMsg && *msgP == "" {
		fatalf("%s: -msg is required", cmd)
	}
	if needLine && *lineP <= 0 {
		fatalf("%s: -line is required (must be >= 1)", cmd)
	}
	return *shardP, *msgP, *lineP, *agentP
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
