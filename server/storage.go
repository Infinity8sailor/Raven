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
	dataDir string
	mu      sync.RWMutex
	store   *TaskStore
}

func NewStorage(dataDir string) *Storage {
	return &Storage{
		dataDir: dataDir,
	}
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
		if err := s.createBackupNoLock(); err != nil {
			log.Printf("[Raven] Error creating backup: %v", err)
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
	return s.createBackupNoLock()
}

func (s *Storage) createBackupNoLock() error {
	src := s.getStorePath()
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	ts := time.Now().Format("2006-01-02_15-04-05.000")
	dst := filepath.Join(s.getBackupsDir(), fmt.Sprintf("tasks_%s.json", ts))

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if s.store != nil {
		s.store.Meta.LastBackup = NowISO()
	}

	if s.store != nil {
		maxCount := s.store.Settings.BackupMaxCount
		if maxCount > 0 {
			files, err := os.ReadDir(s.getBackupsDir())
			if err == nil {
				var backups []os.DirEntry
				for _, f := range files {
					if !f.IsDir() {
						backups = append(backups, f)
					}
				}
				if len(backups) > maxCount {
					sort.Slice(backups, func(i, j int) bool {
						infoI, _ := backups[i].Info()
						infoJ, _ := backups[j].Info()
						if infoI != nil && infoJ != nil {
							return infoI.ModTime().Before(infoJ.ModTime())
						}
						return false
					})
					for i := 0; i < len(backups)-maxCount; i++ {
						os.Remove(filepath.Join(s.getBackupsDir(), backups[i].Name()))
					}
				}
			}
		}
	}
	return nil
}
