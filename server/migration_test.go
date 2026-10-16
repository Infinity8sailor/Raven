package server

import "testing"

func TestMigrateCurrentVersion(t *testing.T) {
	store := &TaskStore{
		SchemaVersion: CurrentSchemaVersion,
		Settings: Settings{
			ActiveProfile: "all",
			Profiles:      []string{"all", "work"},
		},
	}
	migrated, err := Migrate(store)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if migrated {
		t.Error("expected no migration for current version with complete settings")
	}
}

func TestMigrateBackfillProfiles(t *testing.T) {
	store := &TaskStore{
		SchemaVersion: CurrentSchemaVersion,
		Settings:      Settings{},
	}
	migrated, err := Migrate(store)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !migrated {
		t.Error("expected migration when profiles are missing")
	}
	if store.Settings.ActiveProfile != "all" {
		t.Errorf("expected active profile 'all', got '%s'", store.Settings.ActiveProfile)
	}
	if len(store.Settings.Profiles) == 0 {
		t.Error("expected default profiles to be populated")
	}
}

func TestMigrateUnknownVersion(t *testing.T) {
	store := &TaskStore{SchemaVersion: "2.0.0"}
	migrated, err := Migrate(store)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !migrated {
		t.Error("expected migration for future version")
	}
	if store.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected %s, got %s", CurrentSchemaVersion, store.SchemaVersion)
	}
}

func TestMigrateEmptyVersion(t *testing.T) {
	store := &TaskStore{SchemaVersion: ""}
	migrated, err := Migrate(store)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !migrated {
		t.Error("expected migration for empty version")
	}
	if store.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected %s, got %s", CurrentSchemaVersion, store.SchemaVersion)
	}
}
