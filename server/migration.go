package server

import "log"

func Migrate(store *TaskStore) (bool, error) {
	changed := false
	if store.SchemaVersion != CurrentSchemaVersion {
		log.Printf("[Raven] Schema version mismatch: expected %s, got %s", CurrentSchemaVersion, store.SchemaVersion)
		store.SchemaVersion = CurrentSchemaVersion
		store.Meta.LastMigration = CurrentSchemaVersion
		changed = true
	}

	// Ensure Settings has valid profile defaults if missing from older stores
	if store.Settings.ActiveProfile == "" {
		store.Settings.ActiveProfile = "all"
		changed = true
	}
	if len(store.Settings.Profiles) == 0 {
		store.Settings.Profiles = []string{"all", "work", "personal", "ai"}
		changed = true
	}

	return changed, nil
}
