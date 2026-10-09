package server

import (
	"net/http"
	"time"
)

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	store := s.storage.GetStore()
	writeJSON(w, http.StatusOK, store.Settings)
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	if req.Theme != nil {
		valid := false
		for _, t := range ValidThemes {
			if t == *req.Theme {
				valid = true
				break
			}
		}
		if !valid {
			writeError(w, http.StatusBadRequest, "invalid theme", "theme not recognized")
			return
		}
	}

	var updated Settings
	err := s.storage.UpdateStore(func(store *TaskStore) {
		if req.Theme != nil {
			store.Settings.Theme = *req.Theme
		}
		if req.ShowArchived != nil {
			store.Settings.ShowArchived = *req.ShowArchived
		}
		if req.BackupMode != nil {
			store.Settings.BackupMode = *req.BackupMode
		}
		if req.BackupMaxCount != nil {
			store.Settings.BackupMaxCount = *req.BackupMaxCount
		}
		updated = store.Settings
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleBackup(w http.ResponseWriter, r *http.Request) {
	if err := s.storage.CreateBackup(); err != nil {
		writeError(w, http.StatusInternalServerError, "backup failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	store := s.storage.GetStore()
	
	listCount := len(store.Lists)
	itemCount := 0
	for _, l := range store.Lists {
		itemCount += len(l.Items)
	}

	uptime := time.Since(s.startTime).String()

	resp := HealthResponse{
		Status:        "ok",
		SchemaVersion: store.SchemaVersion,
		ListCount:     listCount,
		ItemCount:     itemCount,
		Uptime:        uptime,
	}
	writeJSON(w, http.StatusOK, resp)
}
