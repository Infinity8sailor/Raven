package server

import (
	"net/http"
)

func (s *Server) handleGetLists(w http.ResponseWriter, _ *http.Request) {
	store := s.storage.GetStore()
	showArchived := store.Settings.ShowArchived

	var result []List
	for _, l := range store.Lists {
		if l.Archived && !showArchived {
			continue
		}
		result = append(result, l)
	}

	if result == nil {
		result = []List{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCreateList(w http.ResponseWriter, r *http.Request) {
	var req CreateListRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	if err := ValidateCreateListRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation error", err.Error())
		return
	}

	list := List{
		ID:        GenerateID("lst_"),
		Type:      req.Type,
		Title:     req.Title,
		CreatedAt: NowISO(),
		Features: Features{
			AutoCollapse: true,
		},
		Items: []Item{},
	}

	err := s.storage.UpdateStore(func(store *TaskStore) {
		list.Order = len(store.Lists)
		store.Lists = append(store.Lists, list)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, list)
}

func (s *Server) handleUpdateList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req UpdateListRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	var updated List
	err := s.storage.UpdateStore(func(store *TaskStore) {
		list, _ := findList(store, id)
		if list == nil {
			return
		}
		if req.Title != nil {
			if ValidateListTitle(*req.Title) == nil {
				list.Title = *req.Title
			}
		}
		if req.Pinned != nil {
			list.Pinned = *req.Pinned
		}
		if req.Archived != nil {
			list.Archived = *req.Archived
		}
		if req.SprintStart != nil {
			list.SprintStart = *req.SprintStart
		}
		if req.SprintEnd != nil {
			list.SprintEnd = *req.SprintEnd
		}
		updated = *list
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	if updated.ID == "" {
		writeError(w, http.StatusNotFound, "not found", "")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("confirm") != "true" {
		writeError(w, http.StatusBadRequest, ErrConfirmRequired.Error(), "")
		return
	}

	id := r.PathValue("id")
	found := false
	err := s.storage.UpdateStore(func(store *TaskStore) {
		for i, l := range store.Lists {
			if l.ID == id {
				store.Lists = append(store.Lists[:i], store.Lists[i+1:]...)
				found = true
				break
			}
		}
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleUpdateFeatures(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req UpdateFeaturesRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	var updatedFeatures Features
	found := false
	err := s.storage.UpdateStore(func(store *TaskStore) {
		list, _ := findList(store, id)
		if list == nil {
			return
		}
		found = true
		if req.Reorder != nil {
			list.Features.Reorder = *req.Reorder
		}
		if req.Checkboxes != nil {
			list.Features.Checkboxes = *req.Checkboxes
		}
		if req.DueDate != nil {
			list.Features.DueDate = *req.DueDate
		}
		if req.LastUpdated != nil {
			list.Features.LastUpdated = *req.LastUpdated
		}
		if req.SprintDates != nil {
			list.Features.SprintDates = *req.SprintDates
		}
		if req.AutoCollapse != nil {
			list.Features.AutoCollapse = *req.AutoCollapse
		}
		if req.Sensitive != nil {
			list.Features.Sensitive = *req.Sensitive
		}
		updatedFeatures = list.Features
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found", "")
		return
	}
	writeJSON(w, http.StatusOK, updatedFeatures)
}

func (s *Server) handleReorderLists(w http.ResponseWriter, r *http.Request) {
	var req ReorderRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	err := s.storage.UpdateStore(func(store *TaskStore) {
		for i, l := range store.Lists {
			if l.ID == req.ID {
				if req.Direction == "up" && i > 0 {
					store.Lists[i], store.Lists[i-1] = store.Lists[i-1], store.Lists[i]
					store.Lists[i].Order = i
					store.Lists[i-1].Order = i - 1
				} else if req.Direction == "down" && i < len(store.Lists)-1 {
					store.Lists[i], store.Lists[i+1] = store.Lists[i+1], store.Lists[i]
					store.Lists[i].Order = i
					store.Lists[i+1].Order = i + 1
				}
				break
			}
		}
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
