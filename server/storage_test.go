package server

import (
	"encoding/json"
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
