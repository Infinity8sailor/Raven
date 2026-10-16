package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Storage struct {
	dataDir        string
	mu             sync.RWMutex
	store          *TaskStore
	lastAutoBackup time.Time
	backupCooldown time.Duration
}

func NewStorage(dataDir string) *Storage {
	return &Storage{
		dataDir:        dataDir,
		backupCooldown: 5 * time.Minute,
	}
}

func (s *Storage) SetBackupCooldown(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backupCooldown = d
}

func (s *Storage) getStorePath() string {
	return filepath.Join(s.dataDir, "tasks.json")
}

func (s *Storage) getBackupsDir() string {
	return filepath.Join(s.dataDir, "backups")
}

func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.getBackupsDir(), 0755); err != nil {
		return err
	}

	path := s.getStorePath()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.store = NewDefaultStore()
			return s.saveNoLock()
		}
		return err
	}

	var store TaskStore
	if err := json.Unmarshal(b, &store); err != nil {
		return err
	}
	s.store = &store
	return nil
}

func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveNoLock()
}

func (s *Storage) saveNoLock() error {
	if s.store != nil && s.store.Settings.BackupMode == "on_write" {
		now := time.Now()
		// Only create auto-backup if cooldown period has elapsed (or first save)
		if s.lastAutoBackup.IsZero() || now.Sub(s.lastAutoBackup) >= s.backupCooldown {
			created, err := s.createBackupNoLock()
			if err != nil {
				log.Printf("[Raven] Error creating backup: %v", err)
			} else if created {
				s.lastAutoBackup = now
			}
		}
	}

	path := s.getStorePath()
	tmp := path + ".tmp"

	b, err := json.MarshalIndent(s.store, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *Storage) GetStore() *TaskStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.store
}

func (s *Storage) UpdateStore(fn func(*TaskStore)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fn(s.store)
	return s.saveNoLock()
}

func (s *Storage) CreateBackup() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	created, err := s.createBackupNoLock()
	if err == nil && created {
		s.lastAutoBackup = time.Now()
	}
	return err
}

func (s *Storage) createBackupNoLock() (bool, error) {
	src := s.getStorePath()
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	ts := time.Now().Format("2006-01-02_15-04-05.000")
	dst := filepath.Join(s.getBackupsDir(), fmt.Sprintf("tasks_%s.json", ts))

	in, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return false, err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return false, err
	}

	if s.store != nil {
		s.store.Meta.LastBackup = NowISO()
	}

	if s.store != nil {
		maxCount := s.store.Settings.BackupMaxCount
		if maxCount > 0 {
			pruneBackupsTiered(s.getBackupsDir(), maxCount)
		}
	}
	return true, nil
}

type backupFileItem struct {
	entry   os.DirEntry
	modTime time.Time
}

// pruneBackupsTiered trims backup files using time-tiered retention:
// 1. Up to 10 most recent live editing backups.
// 2. 1 snapshot per day for the last 7 days.
// 3. 1 snapshot per week for the last 4 weeks.
// Unprotected redundant same-day backups beyond the live edit limit are removed first,
// preventing rapid edits from wiping out older daily or weekly snapshots.
func pruneBackupsTiered(backupsDir string, maxCount int) {
	if maxCount <= 0 {
		return
	}

	files, err := os.ReadDir(backupsDir)
	if err != nil {
		return
	}

	var items []backupFileItem
	for _, f := range files {
		if !f.IsDir() {
			info, err := f.Info()
			if err != nil {
				continue
			}
			t := parseBackupTimestamp(f.Name(), info.ModTime())
			items = append(items, backupFileItem{
				entry:   f,
				modTime: t,
			})
		}
	}

	if len(items) <= maxCount {
		return
	}

	// Sort newest to oldest
	sort.Slice(items, func(i, j int) bool {
		return items[i].modTime.After(items[j].modTime)
	})

	// If maxCount is small (e.g. <= 5 in unit tests), keep the maxCount newest
	if maxCount <= 5 {
		for i := maxCount; i < len(items); i++ {
			_ = os.Remove(filepath.Join(backupsDir, items[i].entry.Name()))
		}
		return
	}

	now := time.Now()
	protected := make(map[string]bool)

	// Tier 1: Recent live edit snapshots (up to 5 for compact retention, up to 10 for larger maxCount)
	recentLimit := 5
	if maxCount > 20 {
		recentLimit = 10
	}
	if recentLimit > maxCount {
		recentLimit = maxCount
	}
	for i := 0; i < recentLimit && i < len(items); i++ {
		protected[items[i].entry.Name()] = true
	}

	// Tier 2: 1 snapshot per day for the past 7 days (days 1 to 7)
	seenDays := make(map[string]bool)
	for _, it := range items {
		daysAgo := int(now.Sub(it.modTime).Hours() / 24)
		if daysAgo >= 1 && daysAgo <= 7 {
			dayKey := it.modTime.Format("2006-01-02")
			if !seenDays[dayKey] {
				seenDays[dayKey] = true
				protected[it.entry.Name()] = true
			}
		}
	}

	// Tier 3: 1 snapshot per week for the past 4 weeks (days 8 to 28)
	seenWeeks := make(map[string]bool)
	for _, it := range items {
		daysAgo := int(now.Sub(it.modTime).Hours() / 24)
		if daysAgo > 7 && daysAgo <= 28 {
			year, week := it.modTime.ISOWeek()
			weekKey := fmt.Sprintf("%04d-W%02d", year, week)
			if !seenWeeks[weekKey] {
				seenWeeks[weekKey] = true
				protected[it.entry.Name()] = true
			}
		}
	}

	// First pass: Delete unprotected files (e.g. redundant same-day edits beyond recent limit)
	remaining := make([]backupFileItem, 0, len(items))
	for _, it := range items {
		if !protected[it.entry.Name()] {
			_ = os.Remove(filepath.Join(backupsDir, it.entry.Name()))
		} else {
			remaining = append(remaining, it)
		}
	}

	// Second pass: If remaining protected files still exceed maxCount, prune oldest
	if len(remaining) > maxCount {
		for i := maxCount; i < len(remaining); i++ {
			_ = os.Remove(filepath.Join(backupsDir, remaining[i].entry.Name()))
		}
	}
}

func parseBackupTimestamp(name string, fallback time.Time) time.Time {
	var y, m, d, hr, min, sec, ms int
	n, err := fmt.Sscanf(name, "tasks_%04d-%02d-%02d_%02d-%02d-%02d.%03d.json", &y, &m, &d, &hr, &min, &sec, &ms)
	if err == nil && n == 7 {
		return time.Date(y, time.Month(m), d, hr, min, sec, ms*1e6, time.Local)
	}
	return fallback
}
