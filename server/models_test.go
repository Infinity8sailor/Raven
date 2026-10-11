package server

import (
	"strings"
	"testing"
	"time"
)

func TestValidateListTitle(t *testing.T) {
	if err := ValidateListTitle(""); err == nil {
		t.Error("expected error for empty title")
	}
	if err := ValidateListTitle("a"); err != nil {
		t.Errorf("unexpected error for 1 char: %v", err)
	}
	if err := ValidateListTitle(strings.Repeat("a", 48)); err != nil {
		t.Errorf("unexpected error for 48 chars: %v", err)
	}
	if err := ValidateListTitle(strings.Repeat("a", 49)); err == nil {
		t.Error("expected error for 49 chars")
	}
}

func TestValidateItemTitle(t *testing.T) {
	if err := ValidateItemTitle(""); err == nil {
		t.Error("expected error for empty title")
	}
	if err := ValidateItemTitle("a"); err != nil {
		t.Errorf("unexpected error for 1 char: %v", err)
	}
	if err := ValidateItemTitle(strings.Repeat("a", 200)); err != nil {
		t.Errorf("unexpected error for 200 chars: %v", err)
	}
	if err := ValidateItemTitle(strings.Repeat("a", 201)); err == nil {
		t.Error("expected error for 201 chars")
	}
}

func TestValidateClipboardKey(t *testing.T) {
	if err := ValidateClipboardKey(""); err == nil {
		t.Error("expected error for empty key")
	}
	if err := ValidateClipboardKey("a"); err != nil {
		t.Errorf("unexpected error for 1 char: %v", err)
	}
	if err := ValidateClipboardKey(strings.Repeat("a", 48)); err != nil {
		t.Errorf("unexpected error for 48 chars: %v", err)
	}
	if err := ValidateClipboardKey(strings.Repeat("a", 49)); err == nil {
		t.Error("expected error for 49 chars")
	}
}

func TestValidateClipboardValue(t *testing.T) {
	if err := ValidateClipboardValue(""); err != nil {
		t.Errorf("unexpected error for empty value: %v", err)
	}
	if err := ValidateClipboardValue(strings.Repeat("a", 500)); err != nil {
		t.Errorf("unexpected error for 500 chars: %v", err)
	}
	if err := ValidateClipboardValue(strings.Repeat("a", 501)); err == nil {
		t.Error("expected error for 501 chars")
	}
}

func TestValidateListType(t *testing.T) {
	if err := ValidateListType("standard"); err != nil {
		t.Errorf("unexpected error for standard: %v", err)
	}
	if err := ValidateListType("clipboard"); err != nil {
		t.Errorf("unexpected error for clipboard: %v", err)
	}
	if err := ValidateListType("unknown"); err == nil {
		t.Error("expected error for unknown type")
	}
	if err := ValidateListType(""); err == nil {
		t.Error("expected error for empty type")
	}
}

func TestValidateCreateListRequest(t *testing.T) {
	if err := ValidateCreateListRequest(CreateListRequest{Type: "standard", Title: "title"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ValidateCreateListRequest(CreateListRequest{Type: "clipboard", Title: "title"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ValidateCreateListRequest(CreateListRequest{Type: "standard", Title: ""}); err == nil {
		t.Error("expected error for empty title")
	}
	if err := ValidateCreateListRequest(CreateListRequest{Type: "standard", Title: strings.Repeat("a", 49)}); err == nil {
		t.Error("expected error for long title")
	}
	if err := ValidateCreateListRequest(CreateListRequest{Type: "unknown", Title: "title"}); err == nil {
		t.Error("expected error for bad type")
	}
}

func TestValidateCreateItemRequest(t *testing.T) {
	if err := ValidateCreateItemRequest(CreateItemRequest{Title: "title"}, "standard"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ValidateCreateItemRequest(CreateItemRequest{Title: ""}, "standard"); err == nil {
		t.Error("expected error for empty title")
	}
	if err := ValidateCreateItemRequest(CreateItemRequest{Title: strings.Repeat("a", 201)}, "standard"); err == nil {
		t.Error("expected error for long title")
	}

	if err := ValidateCreateItemRequest(CreateItemRequest{Key: "key", Value: "value"}, "clipboard"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ValidateCreateItemRequest(CreateItemRequest{Key: "", Value: "value"}, "clipboard"); err == nil {
		t.Error("expected error for empty key")
	}
	if err := ValidateCreateItemRequest(CreateItemRequest{Key: strings.Repeat("a", 49), Value: "value"}, "clipboard"); err == nil {
		t.Error("expected error for long key")
	}
	if err := ValidateCreateItemRequest(CreateItemRequest{Key: "key", Value: strings.Repeat("a", 501)}, "clipboard"); err == nil {
		t.Error("expected error for long value")
	}
}

func TestGenerateID(t *testing.T) {
	prefix := "test_"
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateID(prefix)
		if !strings.HasPrefix(id, prefix) {
			t.Errorf("id %s does not have prefix %s", id, prefix)
		}
		if len(id) <= len(prefix) {
			t.Errorf("id %s is too short", id)
		}
		if ids[id] {
			t.Errorf("duplicate id generated: %s", id)
		}
		ids[id] = true
	}
}

func TestNowISO(t *testing.T) {
	iso := NowISO()
	_, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		t.Errorf("failed to parse NowISO() as RFC3339: %v", err)
	}
}

func TestNewDefaultStore(t *testing.T) {
	store := NewDefaultStore()
	if store.SchemaVersion != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", store.SchemaVersion)
	}
	if store.Settings.Theme != "obsidian" {
		t.Errorf("expected obsidian, got %s", store.Settings.Theme)
	}
	if len(store.Lists) == 0 {
		t.Fatalf("expected seed lists, got 0")
	}

	hasStandard := false
	hasClipboard := false
	for _, l := range store.Lists {
		if l.Type == ListTypeStandard {
			hasStandard = true
			if len(l.Items) == 0 {
				t.Errorf("standard list %s should have seed items", l.Title)
			}
		}
		if l.Type == ListTypeClipboard {
			hasClipboard = true
			if len(l.Items) == 0 {
				t.Errorf("clipboard list %s should have seed items", l.Title)
			}
		}
	}
	if !hasStandard || !hasClipboard {
		t.Errorf("expected both standard and clipboard lists, got standard=%v, clipboard=%v", hasStandard, hasClipboard)
	}
}
