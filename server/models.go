package server

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// --- Schema Version ---

const CurrentSchemaVersion = "1.0.0"

// --- Text Length Limits ---
// Enforced at API validation layer. UI truncates visually via CSS.
// Designed for 2K screens (2560px) with 3-column fluid layout.
// 1 line ≈ 100 chars at 14px sans-serif on an ~800px column.

const (
	MaxListTitleLen     = 48  // 1 line, truncate with ellipsis
	MaxItemTitleLen     = 200 // 2 lines max, truncate with ellipsis
	MaxClipboardKeyLen  = 48  // 1 line, truncate with ellipsis
	MaxClipboardValueLen = 500 // 1 line in UI (truncated), generous for URLs/values
)

// --- List Types ---

const (
	ListTypeStandard  = "standard"
	ListTypeClipboard = "clipboard"
)

// --- Item Status ---

const (
	StatusTodo = "todo"
	StatusDone = "done"
)

// --- Errors ---

var (
	ErrNotFound       = errors.New("not found")
	ErrConfirmRequired = errors.New("confirm required")
	ErrValidation     = errors.New("validation error")
)

// --- Core Data Types ---

// TaskStore is the root data structure persisted to tasks.json.
type TaskStore struct {
	SchemaVersion string   `json:"schema_version"`
	Lists         []List   `json:"lists"`
	Settings      Settings `json:"settings"`
	Meta          Meta     `json:"meta"`
}

// List represents a single list — either standard (todo items) or clipboard (key-value pairs).
type List struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Profile     string   `json:"profile,omitempty"`
	CreatedAt   string   `json:"created_at"`
	Order       int      `json:"order"`
	Pinned      bool     `json:"pinned"`
	Archived    bool     `json:"archived"`
	Features    Features `json:"features"`
	SprintStart string   `json:"sprint_start,omitempty"`
	SprintEnd   string   `json:"sprint_end,omitempty"`
	Items       []Item   `json:"items"`
}

// Item represents a single entry in a list.
// For standard lists: Title + Status are used.
// For clipboard lists: Key + Value are used.
type Item struct {
	ID          string `json:"id"`
	Title       string `json:"title,omitempty"`
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	Status      string `json:"status,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Order       int    `json:"order"`
	Collapsed   bool   `json:"collapsed"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// Features controls which capabilities are enabled on a list.
type Features struct {
	Reorder      bool `json:"reorder"`
	Checkboxes   bool `json:"checkboxes"`
	DueDate      bool `json:"due_date"`
	LastUpdated  bool `json:"last_updated"`
	SprintDates  bool `json:"sprint_dates"`
	AutoCollapse bool `json:"auto_collapse"`
	Sensitive    bool `json:"sensitive"`
}

// Settings holds user-configurable app settings.
type Settings struct {
	Theme          string   `json:"theme"`
	ShowArchived   bool     `json:"show_archived"`
	BackupMode     string   `json:"backup_mode"`
	BackupMaxCount int      `json:"backup_max_count"`
	ActiveProfile  string   `json:"active_profile,omitempty"`
	Profiles       []string `json:"profiles,omitempty"`
}

// Meta holds internal bookkeeping data.
type Meta struct {
	LastBackup    string `json:"last_backup"`
	LastMigration string `json:"last_migration"`
}

// --- API Request/Response Types ---

// CreateListRequest is the payload for POST /api/lists.
type CreateListRequest struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Profile string `json:"profile,omitempty"`
}

// UpdateListRequest is the payload for PUT /api/lists/:id.
type UpdateListRequest struct {
	Title       *string `json:"title,omitempty"`
	Profile     *string `json:"profile,omitempty"`
	Pinned      *bool   `json:"pinned,omitempty"`
	Archived    *bool   `json:"archived,omitempty"`
	SprintStart *string `json:"sprint_start,omitempty"`
	SprintEnd   *string `json:"sprint_end,omitempty"`
}

// UpdateFeaturesRequest is the payload for PATCH /api/lists/:id/features.
type UpdateFeaturesRequest struct {
	Reorder      *bool `json:"reorder,omitempty"`
	Checkboxes   *bool `json:"checkboxes,omitempty"`
	DueDate      *bool `json:"due_date,omitempty"`
	LastUpdated  *bool `json:"last_updated,omitempty"`
	SprintDates  *bool `json:"sprint_dates,omitempty"`
	AutoCollapse *bool `json:"auto_collapse,omitempty"`
	Sensitive    *bool `json:"sensitive,omitempty"`
}

// CreateItemRequest is the payload for POST /api/lists/:id/items.
type CreateItemRequest struct {
	Title     string `json:"title,omitempty"`
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
	Sensitive *bool  `json:"sensitive,omitempty"`
}

// UpdateItemRequest is the payload for PUT /api/lists/:id/items/:itemId.
type UpdateItemRequest struct {
	Title     *string `json:"title,omitempty"`
	Key       *string `json:"key,omitempty"`
	Value     *string `json:"value,omitempty"`
	DueDate   *string `json:"due_date,omitempty"`
	Sensitive *bool   `json:"sensitive,omitempty"`
}

// ReorderRequest is the payload for reorder endpoints.
type ReorderRequest struct {
	ID        string `json:"id"`
	Direction string `json:"direction"` // "up" or "down"
}

// UpdateSettingsRequest is the payload for PUT /api/settings.
type UpdateSettingsRequest struct {
	Theme          *string   `json:"theme,omitempty"`
	ShowArchived   *bool     `json:"show_archived,omitempty"`
	BackupMode     *string   `json:"backup_mode,omitempty"`
	BackupMaxCount *int      `json:"backup_max_count,omitempty"`
	ActiveProfile  *string   `json:"active_profile,omitempty"`
	Profiles       *[]string `json:"profiles,omitempty"`
}

// APIError is the standard error response.
type APIError struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// HealthResponse is returned by GET /api/health.
type HealthResponse struct {
	Status        string `json:"status"`
	Version       string `json:"version,omitempty"`
	GitCommit     string `json:"git_commit,omitempty"`
	BuildDate     string `json:"build_date,omitempty"`
	SchemaVersion string `json:"schema_version"`
	ListCount     int    `json:"list_count"`
	ItemCount     int    `json:"item_count"`
	Uptime        string `json:"uptime"`
}

// --- ID Generation ---

// GenerateID creates a random hex ID with the given prefix (e.g., "lst_", "itm_").
func GenerateID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use timestamp-based ID if crypto/rand fails
		return prefix + hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return prefix + hex.EncodeToString(b)
}

// --- Validation ---

// ValidateListTitle checks the list title against length constraints.
func ValidateListTitle(title string) error {
	if len(title) == 0 {
		return errors.New("list title cannot be empty")
	}
	if len(title) > MaxListTitleLen {
		return errors.New("list title exceeds maximum length of 48 characters")
	}
	return nil
}

// ValidateItemTitle checks the standard item title against length constraints.
func ValidateItemTitle(title string) error {
	if len(title) == 0 {
		return errors.New("item title cannot be empty")
	}
	if len(title) > MaxItemTitleLen {
		return errors.New("item title exceeds maximum length of 200 characters")
	}
	return nil
}

// ValidateClipboardKey checks the clipboard key against length constraints.
func ValidateClipboardKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if len(trimmed) == 0 {
		return errors.New("clipboard key cannot be empty")
	}
	if len(trimmed) > MaxClipboardKeyLen {
		return errors.New("clipboard key exceeds maximum length of 48 characters")
	}
	return nil
}

// ValidateClipboardValue checks the clipboard value against length constraints.
func ValidateClipboardValue(value string) error {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) == 0 {
		return errors.New("clipboard value cannot be empty")
	}
	if len(trimmed) > MaxClipboardValueLen {
		return errors.New("clipboard value exceeds maximum length of 500 characters")
	}
	return nil
}

// ValidateListType checks that the list type is valid.
func ValidateListType(listType string) error {
	if listType != ListTypeStandard && listType != ListTypeClipboard {
		return errors.New("list type must be 'standard' or 'clipboard'")
	}
	return nil
}

// ValidateCreateListRequest validates the full create-list payload.
func ValidateCreateListRequest(req CreateListRequest) error {
	if err := ValidateListType(req.Type); err != nil {
		return err
	}
	return ValidateListTitle(req.Title)
}

// ValidateCreateItemRequest validates a new item payload against the parent list type.
func ValidateCreateItemRequest(req CreateItemRequest, listType string) error {
	switch listType {
	case ListTypeStandard:
		return ValidateItemTitle(req.Title)
	case ListTypeClipboard:
		if err := ValidateClipboardKey(req.Key); err != nil {
			return err
		}
		return ValidateClipboardValue(req.Value)
	default:
		return errors.New("unknown list type")
	}
}

// --- Default Data ---

// GetDefaultSeedLists returns introductory lists showcasing the app's features.
func GetDefaultSeedLists() []List {
	now := NowISO()
	return []List{
		{
			ID:        "lst_welcome",
			Type:      ListTypeStandard,
			Title:     "Welcome to Raven",
			CreatedAt: now,
			Order:     0,
			Pinned:    true,
			Archived:  false,
			Features: Features{
				Reorder:      true,
				Checkboxes:   true,
				DueDate:      false,
				LastUpdated:  true,
				SprintDates:  false,
				AutoCollapse: true,
			},
			Items: []Item{
				{
					ID:        "itm_seed_1",
					Title:     "Double-click any item to edit inline (Enter to save)",
					Status:    StatusTodo,
					Order:     0,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "itm_seed_2",
					Title:     "Click checkbox on the left to mark tasks completed",
					Status:    StatusTodo,
					Order:     1,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "itm_seed_3",
					Title:     "Click 3-dots (⋮) to configure features, pin, or archive",
					Status:    StatusTodo,
					Order:     2,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:          "itm_seed_4",
					Title:       "Use ↑ / ↓ arrow buttons to reorder your tasks",
					Status:      StatusDone,
					Order:       3,
					CreatedAt:   now,
					UpdatedAt:   now,
					CompletedAt: now,
				},
				{
					ID:          "itm_seed_5",
					Title:       "Customize themes & fonts in top-right Settings (⛭)",
					Status:      StatusDone,
					Order:       4,
					CreatedAt:   now,
					UpdatedAt:   now,
					CompletedAt: now,
				},
			},
		},
		{
			ID:          "lst_sprint_demo",
			Type:        ListTypeStandard,
			Title:       "Sprint 1: Explore Features",
			CreatedAt:   now,
			Order:       1,
			Pinned:      false,
			Archived:    false,
			SprintStart: "2026-10-01",
			SprintEnd:   "2026-10-15",
			Features: Features{
				Reorder:      true,
				Checkboxes:   true,
				DueDate:      true,
				LastUpdated:  true,
				SprintDates:  true,
				AutoCollapse: true,
			},
			Items: []Item{
				{
					ID:        "itm_seed_6",
					Title:     "Press '/' or click search to filter lists and clipboard keys",
					Status:    StatusTodo,
					DueDate:   "2026-10-12",
					Order:     0,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "itm_seed_7",
					Title:     "Click '+ List' in header to create new task boards",
					Status:    StatusTodo,
					DueDate:   "2026-10-14",
					Order:     1,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:          "itm_seed_8",
					Title:       "Offline JSON storage persists instantly to data/tasks.json",
					Status:      StatusDone,
					Order:       2,
					CreatedAt:   now,
					UpdatedAt:   now,
					CompletedAt: now,
				},
				{
					ID:          "itm_seed_9",
					Title:       "Automated rolling backups kept safely in data/backups",
					Status:      StatusDone,
					Order:       3,
					CreatedAt:   now,
					UpdatedAt:   now,
					CompletedAt: now,
				},
			},
		},
		{
			ID:        "lst_clipboard_demo",
			Type:      ListTypeClipboard,
			Title:     "Clipboard",
			CreatedAt: now,
			Order:     2,
			Pinned:    false,
			Archived:  false,
			Features: Features{
				Sensitive:    true,
				AutoCollapse: true,
			},
			Items: []Item{
				{
					ID:        "itm_seed_10",
					Key:       "Local Dev Server",
					Value:     "http://localhost:9471",
					Order:     0,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "itm_seed_11",
					Key:       "Example Sensitive Entry",
					Value:     "demo_hidden_entry_9471",
					Sensitive: true,
					Order:     1,
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "itm_seed_12",
					Key:       "Quick Search Tip",
					Value:     "Press / anytime to search",
					Order:     2,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		},
	}
}

// NewDefaultStore creates a fresh TaskStore with default settings and introductory seed lists.
func NewDefaultStore() *TaskStore {
	return &TaskStore{
		SchemaVersion: CurrentSchemaVersion,
		Lists:         GetDefaultSeedLists(),
		Settings: Settings{
			Theme:          "obsidian",
			ShowArchived:   false,
			BackupMode:     "on_write",
			BackupMaxCount: 15,
			ActiveProfile:  "all",
			Profiles:       []string{"all", "work", "personal", "ai"},
		},
		Meta: Meta{
			LastBackup:    "",
			LastMigration: CurrentSchemaVersion,
		},
	}
}

// NowISO returns the current time in ISO 8601 format with timezone.
func NowISO() string {
	return time.Now().Format(time.RFC3339)
}

// --- Valid Themes ---

var ValidThemes = []string{"obsidian", "chlorophyll", "infrared", "ultraviolet", "sepia", "random"}
