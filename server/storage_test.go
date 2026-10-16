package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	if s.dataDir != dir {
		t.Errorf("expected %s, got %s", dir, s.dataDir)
	}
}

func TestLoadCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	err := s.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	store := s.GetStore()
	if store == nil {
		t.Fatal("store is nil")
	}
	if store.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected %s, got %s", CurrentSchemaVersion, store.SchemaVersion)
	}

	if _, err := os.Stat(filepath.Join(dir, "tasks.json")); os.IsNotExist(err) {
		t.Error("expected tasks.json to be created")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	s1 := NewStorage(dir)
	s1.Load()
	
	s1.UpdateStore(func(store *TaskStore) {
		store.Lists = []List{{ID: "lst_1", Title: "title1"}}
	})

	s2 := NewStorage(dir)
	err := s2.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	store := s2.GetStore()
	if len(store.Lists) != 1 || store.Lists[0].ID != "lst_1" {
		t.Errorf("failed to load saved data: %+v", store.Lists)
	}
}

func TestLoadCreatesSeedListsWithAppFeatures(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	if err := s.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	store := s.GetStore()
	if len(store.Lists) == 0 {
		t.Fatal("expected seed lists on fresh load, got 0")
	}

	// Verify standard lists with inline edit / completion / 3-dot tips
	var foundWelcome, foundClipboard bool
	for _, l := range store.Lists {
		if l.ID == "lst_welcome" {
			foundWelcome = true
			if !l.Pinned {
				t.Error("welcome list should be pinned")
			}
		}
		if l.Type == ListTypeClipboard {
			foundClipboard = true
			if !l.Features.Sensitive {
				t.Error("sample clipboard list should have sensitive feature enabled")
			}
		}
	}
	if !foundWelcome || !foundClipboard {
		t.Errorf("missing expected seed lists: welcome=%v, clipboard=%v", foundWelcome, foundClipboard)
	}
}

func TestSaveAtomicity(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	s.Load()

	s.UpdateStore(func(store *TaskStore) {
		store.Lists = append(store.Lists, List{ID: "lst_1", Title: "title1"})
	})

	if _, err := os.Stat(filepath.Join(dir, "tasks.json")); os.IsNotExist(err) {
		t.Error("tasks.json should exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "tasks.json.tmp")); !os.IsNotExist(err) {
		t.Error("tasks.json.tmp should not exist after successful save")
	}
}

func TestSaveCreatesBackup(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	s.Load()

	s.UpdateStore(func(store *TaskStore) {
		store.Settings.BackupMode = "on_write"
		store.Lists = append(store.Lists, List{ID: "lst_1"})
	})

	backupsDir := filepath.Join(dir, "backups")
	files, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected backup file to be created")
	}
}

func TestBackupRotation(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	s.Load()

	s.UpdateStore(func(store *TaskStore) {
		store.Settings.BackupMode = "manual" // manually trigger to control timing if needed
		store.Settings.BackupMaxCount = 3
	})

	for i := 0; i < 5; i++ {
		s.CreateBackup()
		time.Sleep(2 * time.Millisecond)
	}

	backupsDir := filepath.Join(dir, "backups")
	files, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 backups, got %d", len(files))
	}
}

func TestLoadCorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	os.WriteFile(path, []byte("{invalid json"), 0644)

	s := NewStorage(dir)
	err := s.Load()
	if err == nil {
		t.Error("expected error loading corrupt JSON")
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	s.Load()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.UpdateStore(func(store *TaskStore) {
				store.Settings.Theme = "theme"
			})
			_ = s.GetStore()
		}(i)
	}
	wg.Wait()
}

func TestUpdateStore(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	s.Load()

	err := s.UpdateStore(func(store *TaskStore) {
		store.Settings.Theme = "chlorophyll"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	store := s.GetStore()
	if store.Settings.Theme != "chlorophyll" {
		t.Errorf("expected chlorophyll, got %s", store.Settings.Theme)
	}
}

func TestCreateBackup(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	s.Load()

	err := s.CreateBackup()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	backupsDir := filepath.Join(dir, "backups")
	files, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 backup, got %d", len(files))
	}

	b, _ := os.ReadFile(filepath.Join(backupsDir, files[0].Name()))
	var store TaskStore
	if err := json.Unmarshal(b, &store); err != nil {
		t.Errorf("failed to parse backup json: %v", err)
	}
}

func TestBackupCooldown(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)
	s.Load()
	s.SetBackupCooldown(1 * time.Hour) // 1-hour cooldown

	// First mutation creates initial backup
	s.UpdateStore(func(store *TaskStore) {
		store.Settings.BackupMode = "on_write"
		store.Lists = append(store.Lists, List{ID: "lst_1", Title: "L1"})
	})

	// Immediate second mutation should be debounced by cooldown
	s.UpdateStore(func(store *TaskStore) {
		store.Lists = append(store.Lists, List{ID: "lst_2", Title: "L2"})
	})

	backupsDir := filepath.Join(dir, "backups")
	files, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 debounced backup, got %d", len(files))
	}

	// Manual CreateBackup should bypass cooldown
	if err := s.CreateBackup(); err != nil {
		t.Fatalf("unexpected error on manual backup: %v", err)
	}

	filesAfter, _ := os.ReadDir(backupsDir)
	if len(filesAfter) != 2 {
		t.Errorf("expected manual backup to bypass cooldown, got %d files", len(filesAfter))
	}
}

func TestBackupTieredRetention(t *testing.T) {
	dir := t.TempDir()
	backupsDir := filepath.Join(dir, "backups")
	os.MkdirAll(backupsDir, 0755)

	now := time.Now()

	// Create 15 mock backups today
	for i := 0; i < 15; i++ {
		tStamp := now.Add(-time.Duration(i) * time.Minute)
		fn := fmt.Sprintf("tasks_%s.json", tStamp.Format("2006-01-02_15-04-05.000"))
		os.WriteFile(filepath.Join(backupsDir, fn), []byte("{}"), 0644)
	}

	// Create 1 backup yesterday (24h ago)
	yesterday := now.Add(-24 * time.Hour)
	fnYesterday := fmt.Sprintf("tasks_%s.json", yesterday.Format("2006-01-02_15-04-05.000"))
	os.WriteFile(filepath.Join(backupsDir, fnYesterday), []byte("{}"), 0644)

	// Create 1 backup 3 days ago
	threeDaysAgo := now.Add(-72 * time.Hour)
	fn3Days := fmt.Sprintf("tasks_%s.json", threeDaysAgo.Format("2006-01-02_15-04-05.000"))
	os.WriteFile(filepath.Join(backupsDir, fn3Days), []byte("{}"), 0644)

	// Create 1 backup 14 days ago (weekly tier)
	twoWeeksAgo := now.Add(-14 * 24 * time.Hour)
	fn2Weeks := fmt.Sprintf("tasks_%s.json", twoWeeksAgo.Format("2006-01-02_15-04-05.000"))
	os.WriteFile(filepath.Join(backupsDir, fn2Weeks), []byte("{}"), 0644)

	// Total files created = 15 + 1 + 1 + 1 = 18
	// Prune with maxCount = 12 (compact mode uses recentLimit = 5)
	// Expected: 5 newest from today + 1 yesterday + 1 three days ago + 1 two weeks ago = 8 preserved
	// The 10 older same-day edits should be pruned!
	pruneBackupsTiered(backupsDir, 12)

	remaining, _ := os.ReadDir(backupsDir)
	if len(remaining) != 8 {
		t.Fatalf("expected 8 remaining files, got %d", len(remaining))
	}

	// Verify yesterday's file and 3 days ago's file were protected and still exist!
	if _, err := os.Stat(filepath.Join(backupsDir, fnYesterday)); os.IsNotExist(err) {
		t.Errorf("yesterday's backup should have been protected")
	}
	if _, err := os.Stat(filepath.Join(backupsDir, fn3Days)); os.IsNotExist(err) {
		t.Errorf("3 days ago backup should have been protected")
	}
}
