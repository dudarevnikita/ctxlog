// Package memory implements sharded append-only JSONL storage
// with BSD flock for cross-process safety.
package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Entry is a single JSONL record.
type Entry struct {
	Ts    int64  `json:"ts"`
	Agent string `json:"agent"`
	Msg   string `json:"msg"`
}

// Store manages the on-disk .ctxlog directory.
type Store struct {
	basePath string
}

// NewStore creates a Store rooted at basePath/.ctxlog.
func NewStore(basePath string) *Store {
	return &Store{basePath: filepath.Join(basePath, ".ctxlog")}
}

// Init ensures the base .ctxlog directory exists.
func (s *Store) Init() error {
	return os.MkdirAll(s.basePath, 0o755)
}

func (s *Store) shardPath(shard string) string {
	return filepath.Join(s.basePath, shard+".jsonl")
}

// Append writes a single Entry to the named shard.
// Uses exclusive flock to prevent corruption from concurrent writers.
func (s *Store) Append(shard string, e Entry) error {
	if e.Ts == 0 {
		e.Ts = time.Now().Unix()
	}

	path := s.shardPath(shard)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir shard dir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open shard: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// ReadRecent returns the last maxLines entries from the named shard.
// Returns an empty slice if the shard does not exist.
func (s *Store) ReadRecent(shard string, maxLines int) ([]Entry, error) {
	path := s.shardPath(shard)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("open shard: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return nil, fmt.Errorf("flock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	ring := make([]Entry, 0, maxLines)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if len(ring) < maxLines {
			ring = append(ring, e)
		} else {
			copy(ring, ring[1:])
			ring[maxLines-1] = e
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return ring, nil
}
