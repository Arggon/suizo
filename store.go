package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// store persists a Tournament as one JSON document at a fixed path.
// Save is atomic: the document is written to a sibling temp file and renamed,
// so a crash mid-write never leaves a truncated tournament behind.
type store struct {
	path string
}

func openStore(path string) (*store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}
	return &store{path: path}, nil
}

// load reads the tournament document. A missing file is not an error: it
// yields a fresh empty tournament, so the very first command works with no
// setup step.
func (s *store) load() (*Tournament, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &Tournament{Name: "Suizo tournament"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", s.path, err)
	}
	var t Tournament
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	return &t, nil
}

func (s *store) save(t *Tournament) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("encode tournament: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(s.path), filepath.Base(s.path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replace %s: %w", s.path, err)
	}
	return nil
}
