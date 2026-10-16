package server

import (
	"net/http"
)

func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("id")
	var req CreateItemRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	var newItem Item
	var saved bool
	var valErr error
	err := s.storage.UpdateStore(func(store *TaskStore) {
		list, _ := findList(store, listID)
		if list == nil {
			return
		}
		if err := ValidateCreateItemRequest(req, list.Type); err != nil {
			valErr = err
			return
		}
		newItem = Item{
			ID:        GenerateID("itm_"),
			Title:     req.Title,
			Key:       req.Key,
			Value:     req.Value,
			Order:     len(list.Items),
			CreatedAt: NowISO(),
			UpdatedAt: NowISO(),
		}
		if req.Sensitive != nil {
			newItem.Sensitive = *req.Sensitive
		}
		if list.Type == ListTypeStandard {
			newItem.Status = StatusTodo
		}
		list.Items = append(list.Items, newItem)
		saved = true
	})
	if valErr != nil {
		writeError(w, http.StatusBadRequest, "validation error", valErr.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	if !saved {
		writeError(w, http.StatusNotFound, "list not found", "")
		return
	}
	writeJSON(w, http.StatusCreated, newItem)
}

func (s *Server) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("id")
	itemID := r.PathValue("itemId")
	var req UpdateItemRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	var updated Item
	found := false
	err := s.storage.UpdateStore(func(store *TaskStore) {
		list, _ := findList(store, listID)
		if list == nil {
			return
		}
		item, _ := findItem(list, itemID)
		if item == nil {
			return
		}
		found = true
		changed := false
		if req.Title != nil {
			if ValidateItemTitle(*req.Title) == nil {
				item.Title = *req.Title
				changed = true
			}
		}
		if req.Key != nil {
			if ValidateClipboardKey(*req.Key) == nil {
				item.Key = *req.Key
				changed = true
			}
		}
		if req.Value != nil {
			if ValidateClipboardValue(*req.Value) == nil {
				item.Value = *req.Value
				changed = true
			}
		}
		if req.DueDate != nil {
			item.DueDate = *req.DueDate
			changed = true
		}
		if changed {
			item.UpdatedAt = NowISO()
		}
		updated = *item
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found", "")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("confirm") != "true" {
		writeError(w, http.StatusBadRequest, ErrConfirmRequired.Error(), "")
		return
	}
	listID := r.PathValue("id")
	itemID := r.PathValue("itemId")

	found := false
	err := s.storage.UpdateStore(func(store *TaskStore) {
		list, _ := findList(store, listID)
		if list == nil {
			return
		}
		for i, itm := range list.Items {
			if itm.ID == itemID {
				list.Items = append(list.Items[:i], list.Items[i+1:]...)
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

func (s *Server) handleToggleStatus(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("id")
	itemID := r.PathValue("itemId")

	var updated Item
	found := false
	err := s.storage.UpdateStore(func(store *TaskStore) {
		list, _ := findList(store, listID)
		if list == nil {
			return
		}
		item, _ := findItem(list, itemID)
		if item == nil {
			return
		}
		found = true
		if item.Status == StatusTodo {
			item.Status = StatusDone
			item.CompletedAt = NowISO()
		} else {
			item.Status = StatusTodo
			item.CompletedAt = ""
		}
		item.UpdatedAt = NowISO()
		updated = *item
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save", err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "not found", "")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleReorderItems(w http.ResponseWriter, r *http.Request) {
	listID := r.PathValue("id")
	var req ReorderRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	found := false
	err := s.storage.UpdateStore(func(store *TaskStore) {
		list, _ := findList(store, listID)
		if list == nil {
			return
		}
		for i, itm := range list.Items {
			if itm.ID == req.ID {
				found = true
				if req.Direction == "up" && i > 0 {
					list.Items[i], list.Items[i-1] = list.Items[i-1], list.Items[i]
					list.Items[i].Order = i
					list.Items[i-1].Order = i - 1
				} else if req.Direction == "down" && i < len(list.Items)-1 {
					list.Items[i], list.Items[i+1] = list.Items[i+1], list.Items[i]
					list.Items[i].Order = i
					list.Items[i+1].Order = i + 1
				}
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
	writeJSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
