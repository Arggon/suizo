package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// lockPollInterval is how long withLock waits between non-blocking flock
// attempts. A short poll keeps the single-user fast path snappy while still
// converging quickly under contention.
const lockPollInterval = 5 * time.Millisecond

// lockTimeout bounds how long a process will wait for the advisory lock
// before giving up, so a wedged holder cannot hang the CLI forever.
const lockTimeout = 5 * time.Second

// withLock runs fn while holding an exclusive advisory flock on a sidecar
// lockfile next to the store document. Concurrent suizo processes serialize
// their load→mutate→save cycles through this lock, so no write is ever lost
// to a lost-update race. The lock is advisory and stdlib-only
// (syscall.Flock); if the lockfile cannot be opened the error is returned
// rather than silently proceeding unlocked.
func (s *store) withLock(fn func() error) error {
	f, err := os.OpenFile(s.lockPath(), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer func() { _ = f.Close() }()

	deadline := time.Now().Add(lockTimeout)
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) && !errors.Is(err, syscall.EINTR) {
			return fmt.Errorf("lock %s: %w", s.lockPath(), err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for lock %s", s.lockPath())
		}
		time.Sleep(lockPollInterval)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	return fn()
}

func (s *store) lockPath() string {
	return s.path + ".lock"
}

// update performs a locked read-modify-write cycle: it acquires the advisory
// lock, loads the tournament, applies mutate, and saves — all while holding
// the lock, so concurrent invocations cannot overwrite each other's writes.
func (s *store) update(mutate func(t *Tournament) error) error {
	return s.withLock(func() error {
		t, err := s.load()
		if err != nil {
			return err
		}
		if err := mutate(t); err != nil {
			return err
		}
		return s.save(t)
	})
}
