package server

import "testing"

func TestMigrateCurrentVersion(t *testing.T) {
	store := &TaskStore{SchemaVersion: CurrentSchemaVersion}
	migrated, err := Migrate(store)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if migrated {
		t.Error("expected no migration for current version")
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
