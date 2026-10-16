package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	storage := NewStorage(dir)
	storage.Load()
	storage.UpdateStore(func(store *TaskStore) {
		store.Lists = []List{}
	})
	cfg := &Config{Port: 8080, WebDir: t.TempDir()}
	s := NewServer(cfg, storage)
	s.SetupRoutes()
	return s
}

func doRequest(t *testing.T, s *Server, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, "/api"+path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	return rr
}

func TestCreateList(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "Test List"})
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
	var l List
	json.Unmarshal(rr.Body.Bytes(), &l)
	if l.ID == "" || l.Type != "standard" {
		t.Errorf("invalid list returned: %+v", l)
	}
}

func TestCreateListClipboard(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "clipboard", Title: "Clip List"})
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
	var l List
	json.Unmarshal(rr.Body.Bytes(), &l)
	if l.Type != "clipboard" {
		t.Errorf("invalid list returned: %+v", l)
	}
}

func TestCreateListValidation(t *testing.T) {
	s := setupTestServer(t)
	
	rr := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: ""})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty title, got %d", rr.Code)
	}
	
	rr = doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: strings.Repeat("a", 50)})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for long title, got %d", rr.Code)
	}

	rr = doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "invalid", Title: "title"})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad type, got %d", rr.Code)
	}
}

func TestGetLists(t *testing.T) {
	s := setupTestServer(t)
	doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 2"})

	rr := doRequest(t, s, "GET", "/lists", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var lists []List
	json.Unmarshal(rr.Body.Bytes(), &lists)
	if len(lists) != 2 {
		t.Errorf("expected 2 lists, got %d", len(lists))
	}
}

func TestGetListsHidesArchived(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	archived := true
	doRequest(t, s, "PUT", "/lists/"+l.ID, UpdateListRequest{Archived: &archived})

	s.storage.UpdateStore(func(store *TaskStore) {
		store.Settings.ShowArchived = false
	})

	rr2 := doRequest(t, s, "GET", "/lists", nil)
	var lists []List
	json.Unmarshal(rr2.Body.Bytes(), &lists)
	if len(lists) != 0 {
		t.Errorf("expected 0 lists, got %d", len(lists))
	}
}

func TestGetListsShowsArchived(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	archived := true
	doRequest(t, s, "PUT", "/lists/"+l.ID, UpdateListRequest{Archived: &archived})

	s.storage.UpdateStore(func(store *TaskStore) {
		store.Settings.ShowArchived = true
	})

	rr2 := doRequest(t, s, "GET", "/lists", nil)
	var lists []List
	json.Unmarshal(rr2.Body.Bytes(), &lists)
	if len(lists) != 1 {
		t.Errorf("expected 1 lists, got %d", len(lists))
	}
}

func TestUpdateList(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	newTitle := "New Title"
	rr2 := doRequest(t, s, "PUT", "/lists/"+l.ID, UpdateListRequest{Title: &newTitle})
	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}
	var l2 List
	json.Unmarshal(rr2.Body.Bytes(), &l2)
	if l2.Title != "New Title" {
		t.Errorf("expected title to change, got %s", l2.Title)
	}
}

func TestDeleteListWithoutConfirm(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "DELETE", "/lists/id1", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDeleteListWithConfirm(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	rr2 := doRequest(t, s, "DELETE", "/lists/"+l.ID+"?confirm=true", nil)
	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}

	rr3 := doRequest(t, s, "GET", "/lists", nil)
	var lists []List
	json.Unmarshal(rr3.Body.Bytes(), &lists)
	if len(lists) != 0 {
		t.Errorf("expected 0 lists, got %d", len(lists))
	}
}

func TestUpdateFeatures(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	reorder := false
	rr2 := doRequest(t, s, "PATCH", "/lists/"+l.ID+"/features", UpdateFeaturesRequest{Reorder: &reorder})
	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}
	var f Features
	json.Unmarshal(rr2.Body.Bytes(), &f)
	if f.Reorder != false {
		t.Errorf("expected reorder to be false")
	}
}

func TestCreateItem(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	rr2 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: "Item 1"})
	if rr2.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr2.Code)
	}
	var item Item
	json.Unmarshal(rr2.Body.Bytes(), &item)
	if item.Title != "Item 1" {
		t.Errorf("expected Item 1, got %s", item.Title)
	}
}

func TestCreateClipboardItem(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "clipboard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	rr2 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Key: "k1", Value: "v1"})
	if rr2.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr2.Code)
	}
	var item Item
	json.Unmarshal(rr2.Body.Bytes(), &item)
	if item.Key != "k1" || item.Value != "v1" || item.Sensitive {
		t.Errorf("expected k1/v1/non-sensitive, got %s/%s/%v", item.Key, item.Value, item.Sensitive)
	}

	// Test sensitive item with valid key and value
	sens := true
	rr3 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Key: "k2", Value: "v2", Sensitive: &sens})
	if rr3.Code != http.StatusCreated {
		t.Errorf("expected 201 for sensitive key/value, got %d", rr3.Code)
	}
	var item2 Item
	json.Unmarshal(rr3.Body.Bytes(), &item2)
	if item2.Key != "k2" || item2.Value != "v2" || !item2.Sensitive {
		t.Errorf("expected k2/v2 and sensitive=true, got %s/%s/%v", item2.Key, item2.Value, item2.Sensitive)
	}

	// Rejects empty key
	rrEmptyKey := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Key: "", Value: "v3"})
	if rrEmptyKey.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty key, got %d", rrEmptyKey.Code)
	}

	// Rejects empty value
	rrEmptyVal := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Key: "k3", Value: ""})
	if rrEmptyVal.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty value, got %d", rrEmptyVal.Code)
	}

	// Verify that Sensitive cannot be changed via PUT update on existing item (creation only)
	sensFalse := false
	rrUpdate := doRequest(t, s, "PUT", "/lists/"+l.ID+"/items/"+item2.ID, UpdateItemRequest{Sensitive: &sensFalse})
	if rrUpdate.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rrUpdate.Code)
	}
	var updatedItem Item
	json.Unmarshal(rrUpdate.Body.Bytes(), &updatedItem)
	if !updatedItem.Sensitive {
		t.Errorf("sensitive flag must not change on update, should remain true")
	}
}

func TestCreateItemValidation(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	rr2 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: ""})
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty title, got %d", rr2.Code)
	}

	rr3 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: strings.Repeat("a", 250)})
	if rr3.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for long title, got %d", rr3.Code)
	}
}

func TestDeleteItemWithoutConfirm(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "DELETE", "/lists/l1/items/i1", nil)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDeleteItemWithConfirm(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	rr2 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: "Item 1"})
	var item Item
	json.Unmarshal(rr2.Body.Bytes(), &item)

	rr3 := doRequest(t, s, "DELETE", "/lists/"+l.ID+"/items/"+item.ID+"?confirm=true", nil)
	if rr3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr3.Code)
	}
}

func TestToggleStatus(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	rr2 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: "Item 1"})
	var item Item
	json.Unmarshal(rr2.Body.Bytes(), &item)
	if item.Status != StatusTodo {
		t.Errorf("expected todo, got %s", item.Status)
	}

	rr3 := doRequest(t, s, "PATCH", "/lists/"+l.ID+"/items/"+item.ID+"/status", nil)
	if rr3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr3.Code)
	}
	var toggledItem Item
	json.Unmarshal(rr3.Body.Bytes(), &toggledItem)
	if toggledItem.Status != StatusDone || toggledItem.CompletedAt == "" {
		t.Errorf("expected done with completed_at, got %s / %s", toggledItem.Status, toggledItem.CompletedAt)
	}

	rr4 := doRequest(t, s, "PATCH", "/lists/"+l.ID+"/items/"+item.ID+"/status", nil)
	var untoggledItem Item
	json.Unmarshal(rr4.Body.Bytes(), &untoggledItem)
	if untoggledItem.Status != StatusTodo || untoggledItem.CompletedAt != "" {
		t.Errorf("expected todo with no completed_at, got %s / %s", untoggledItem.Status, untoggledItem.CompletedAt)
	}
}

func TestReorderItems(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: "Item 1"})
	rrItem2 := doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: "Item 2"})
	var item2 Item
	json.Unmarshal(rrItem2.Body.Bytes(), &item2)
	doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Title: "Item 3"})

	rrReorder := doRequest(t, s, "POST", "/lists/"+l.ID+"/items/reorder", ReorderRequest{ID: item2.ID, Direction: "up"})
	if rrReorder.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rrReorder.Code)
	}

	// Fetch lists again to verify order
	rr2 := doRequest(t, s, "GET", "/lists", nil)
	var lists []List
	json.Unmarshal(rr2.Body.Bytes(), &lists)
	if len(lists) > 0 && len(lists[0].Items) == 3 {
		if lists[0].Items[0].ID != item2.ID {
			t.Errorf("expected item2 to be first, got %s", lists[0].Items[0].Title)
		}
	}
}

func TestSearch(t *testing.T) {
	s := setupTestServer(t)
	doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "Unique Title"})
	
	rr := doRequest(t, s, "GET", "/search?q=Unique", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	
	var res []interface{}
	json.Unmarshal(rr.Body.Bytes(), &res)
	
	if len(res) == 0 {
		t.Errorf("expected results in search, got %+v", res)
	}
}

func TestSearchSensitive(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "clipboard", Title: "Sensitive List"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	sensitive := true
	doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Key: "k1", Value: "masked_test_entry", Sensitive: &sensitive})

	rr := doRequest(t, s, "GET", "/search?q=masked_test_entry", nil)
	var res []interface{}
	json.Unmarshal(rr.Body.Bytes(), &res)
	
	if len(res) != 0 {
		t.Errorf("expected no items in sensitive search, got %d", len(res))
	}

	// Search for key 'k1': should match and mask sensitive value
	rrKey := doRequest(t, s, "GET", "/search?q=k1", nil)
	var matchedLists []List
	json.Unmarshal(rrKey.Body.Bytes(), &matchedLists)
	if len(matchedLists) == 0 || len(matchedLists[0].Items) == 0 {
		t.Fatalf("expected matched clipboard item for key k1, got none")
	}
	expectedMask := strings.Repeat("•", len("masked_test_entry"))
	if matchedLists[0].Items[0].Value != expectedMask {
		t.Errorf("expected masked value %s for sensitive clipboard item, got %s", expectedMask, matchedLists[0].Items[0].Value)
	}
}

func TestGetSettings(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "GET", "/settings", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var settings Settings
	json.Unmarshal(rr.Body.Bytes(), &settings)
	if settings.Theme != "obsidian" {
		t.Errorf("expected obsidian, got %s", settings.Theme)
	}
}

func TestUpdateSettings(t *testing.T) {
	s := setupTestServer(t)
	theme := "chlorophyll"
	rr := doRequest(t, s, "PUT", "/settings", UpdateSettingsRequest{Theme: &theme})
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var settings Settings
	json.Unmarshal(rr.Body.Bytes(), &settings)
	if settings.Theme != "chlorophyll" {
		t.Errorf("expected chlorophyll, got %s", settings.Theme)
	}
}

func TestUpdateSettingsInvalidTheme(t *testing.T) {
	s := setupTestServer(t)
	theme := "invalid_theme"
	rr := doRequest(t, s, "PUT", "/settings", UpdateSettingsRequest{Theme: &theme})
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHealth(t *testing.T) {
	s := setupTestServer(t)
	s.SetVersionInfo("v1.0.0", "commit123", "2026-10-12T19:00:00Z")
	rr := doRequest(t, s, "GET", "/health", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var h HealthResponse
	json.Unmarshal(rr.Body.Bytes(), &h)
	if h.Status != "ok" || h.Uptime == "" {
		t.Errorf("expected ok with uptime, got %+v", h)
	}
	if h.Version != "v1.0.0" || h.GitCommit != "commit123" {
		t.Errorf("expected v1.0.0 and commit123, got %s / %s", h.Version, h.GitCommit)
	}
}

func TestBackup(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "POST", "/backup", nil)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestPinList(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "Pin Test List"})
	var l List
	json.Unmarshal(rr.Body.Bytes(), &l)

	pinned := true
	rrUpdate := doRequest(t, s, "PUT", "/lists/"+l.ID, UpdateListRequest{Pinned: &pinned})
	if rrUpdate.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrUpdate.Code)
	}
	var updated List
	json.Unmarshal(rrUpdate.Body.Bytes(), &updated)
	if !updated.Pinned {
		t.Errorf("expected list to be pinned")
	}
}

func TestUpdateSprintDates(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "Sprint List"})
	var l List
	json.Unmarshal(rr.Body.Bytes(), &l)

	start := "2026-10-01"
	end := "2026-10-15"
	rrUpdate := doRequest(t, s, "PUT", "/lists/"+l.ID, UpdateListRequest{SprintStart: &start, SprintEnd: &end})
	if rrUpdate.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrUpdate.Code)
	}
	var updated List
	json.Unmarshal(rrUpdate.Body.Bytes(), &updated)
	if updated.SprintStart != "2026-10-01" || updated.SprintEnd != "2026-10-15" {
		t.Errorf("expected sprint dates to update, got %s -> %s", updated.SprintStart, updated.SprintEnd)
	}
}

func TestUpdateAllFeatures(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "Feature Test"})
	var l List
	json.Unmarshal(rr.Body.Bytes(), &l)

	reorder := true
	checkboxes := false
	sprintDates := true
	autoCollapse := false
	rrPatch := doRequest(t, s, "PATCH", "/lists/"+l.ID+"/features", UpdateFeaturesRequest{
		Reorder:      &reorder,
		Checkboxes:   &checkboxes,
		SprintDates:  &sprintDates,
		AutoCollapse: &autoCollapse,
	})
	if rrPatch.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrPatch.Code)
	}
	var updated Features
	json.Unmarshal(rrPatch.Body.Bytes(), &updated)
	if !updated.Reorder || updated.Checkboxes || !updated.SprintDates || updated.AutoCollapse {
		t.Errorf("features mismatch: %+v", updated)
	}
}

func TestReorderList(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 1"})
	var l1 List
	json.Unmarshal(rr1.Body.Bytes(), &l1)

	rr2 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "List 2"})
	var l2 List
	json.Unmarshal(rr2.Body.Bytes(), &l2)

	// Move List 2 up
	rrReorder := doRequest(t, s, "POST", "/lists/reorder", ReorderRequest{ID: l2.ID, Direction: "up"})
	if rrReorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrReorder.Code)
	}

	rrGet := doRequest(t, s, "GET", "/lists", nil)
	var lists []List
	json.Unmarshal(rrGet.Body.Bytes(), &lists)
	if len(lists) < 2 || lists[0].ID != l2.ID {
		t.Errorf("expected List 2 to be first after moving up")
	}
}

func TestSearchKeyOnly_IgnoresValueMatches(t *testing.T) {
	s := setupTestServer(t)
	rr1 := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "clipboard", Title: "Endpoints"})
	var l List
	json.Unmarshal(rr1.Body.Bytes(), &l)

	doRequest(t, s, "POST", "/lists/"+l.ID+"/items", CreateItemRequest{Key: "database_url", Value: "internal_cluster_host_123"})

	// Search for value text: MUST NOT match
	rrVal := doRequest(t, s, "GET", "/search?q=internal_cluster_host", nil)
	var resVal []List
	json.Unmarshal(rrVal.Body.Bytes(), &resVal)
	if len(resVal) != 0 {
		t.Errorf("expected 0 search matches for value text, got %d", len(resVal))
	}

	// Search for key: MUST match
	rrKey := doRequest(t, s, "GET", "/search?q=database", nil)
	var resKey []List
	json.Unmarshal(rrKey.Body.Bytes(), &resKey)
	if len(resKey) == 0 || len(resKey[0].Items) == 0 {
		t.Errorf("expected match for key 'database', got none")
	}
}

func TestCreateListWithProfile(t *testing.T) {
	s := setupTestServer(t)
	rr := doRequest(t, s, "POST", "/lists", CreateListRequest{Type: "standard", Title: "Work Tasks", Profile: "work"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var l List
	json.Unmarshal(rr.Body.Bytes(), &l)
	if l.Profile != "work" {
		t.Errorf("expected profile work, got %s", l.Profile)
	}

	newProf := "ai"
	rrUp := doRequest(t, s, "PUT", "/lists/"+l.ID, UpdateListRequest{Profile: &newProf})
	if rrUp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrUp.Code)
	}
	var updated List
	json.Unmarshal(rrUp.Body.Bytes(), &updated)
	if updated.Profile != "ai" {
		t.Errorf("expected updated profile ai, got %s", updated.Profile)
	}
}

func TestUpdateSettingsProfiles(t *testing.T) {
	s := setupTestServer(t)
	activeProf := "ai"
	profs := []string{"all", "work", "personal", "ai"}
	rr := doRequest(t, s, "PUT", "/settings", UpdateSettingsRequest{
		ActiveProfile: &activeProf,
		Profiles:      &profs,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var st Settings
	json.Unmarshal(rr.Body.Bytes(), &st)
	if st.ActiveProfile != "ai" {
		t.Errorf("expected active profile ai, got %s", st.ActiveProfile)
	}
	if len(st.Profiles) != 4 {
		t.Errorf("expected 4 profiles, got %d", len(st.Profiles))
	}
}
