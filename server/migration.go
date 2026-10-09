package server

import "log"

func Migrate(store *TaskStore) (bool, error) {
	if store.SchemaVersion != CurrentSchemaVersion {
		log.Printf("[Raven] Schema version mismatch: expected %s, got %s", CurrentSchemaVersion, store.SchemaVersion)
		store.SchemaVersion = CurrentSchemaVersion
		store.Meta.LastMigration = CurrentSchemaVersion
		return true, nil
	}
	return false, nil
}
