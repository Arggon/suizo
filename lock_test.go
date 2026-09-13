package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
)

func lockFileExclusive(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// TestStoreUpdateConcurrentNoLostWrites is table-driven: for each scenario it
// spawns N goroutines performing concurrent locked update cycles and asserts
// that every write survived (no lost updates) and ids never collide.
func TestStoreUpdateConcurrentNoLostWrites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		processes int
		existing  int // players already in the store before the storm
	}{
		{name: "two concurrent writers", processes: 2, existing: 0},
		{name: "eight concurrent writers", processes: 8, existing: 3},
		{name: "thirty-two concurrent writers", processes: 32, existing: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "suizo.json")
			s, err := openStore(path)
			if err != nil {
				t.Fatalf("openStore: %v", err)
			}
			for i := 0; i < tc.existing; i++ {
				if err := s.update(func(t *Tournament) error {
					_, err := t.addPlayer(fmt.Sprintf("seed%d", i))
					return err
				}); err != nil {
					t.Fatalf("seed player %d: %v", i, err)
				}
			}

			var wg sync.WaitGroup
			errs := make([]error, tc.processes)
			for i := 0; i < tc.processes; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					errs[i] = s.update(func(t *Tournament) error {
						_, err := t.addPlayer(fmt.Sprintf("P%02d", i))
						return err
					})
				}(i)
			}
			wg.Wait()
			for i, err := range errs {
				if err != nil {
					t.Fatalf("concurrent update %d: %v", i, err)
				}
			}

			final, err := s.load()
			if err != nil {
				t.Fatalf("final load: %v", err)
			}
			want := tc.existing + tc.processes
			if len(final.Players) != want {
				t.Fatalf("lost writes: got %d players, want %d", len(final.Players), want)
			}
			seen := make(map[string]bool)
			for _, p := range final.Players {
				if seen[p.ID] {
					t.Fatalf("duplicate id %q", p.ID)
				}
				seen[p.ID] = true
			}
			for i := 0; i < tc.processes; i++ {
				want := fmt.Sprintf("P%02d", i)
				if !seenPlayer(final, want) {
					t.Fatalf("player %q was lost", want)
				}
			}
		})
	}
}

func seenPlayer(t *Tournament, name string) bool {
	for _, p := range t.Players {
		if p.Name == name {
			return true
		}
	}
	return false
}

// TestRunWithLockTimeout verifies a stuck lock holder surfaces a timeout
// error instead of hanging the CLI forever.
func TestRunWithLockTimeout(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "suizo.json")
	s, err := openStore(path)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	// Hold the lock from this process to simulate a stuck writer.
	f, err := os.OpenFile(s.lockPath(), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("open lock: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	if err := lockFileExclusive(f); err != nil {
		t.Fatalf("lock: %v", err)
	}

	err = s.withLock(func() error { return nil })
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out waiting for lock") {
		t.Fatalf("unexpected error: %v", err)
	}
}
