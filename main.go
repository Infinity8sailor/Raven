package main

import (
	"log"
	"os"
	"path/filepath"

	"raven/server"
)

func main() {
	// If current working directory lacks web assets, switch to executable directory
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		if _, err := os.Stat("web"); os.IsNotExist(err) {
			if _, err := os.Stat(filepath.Join(exeDir, "web")); err == nil {
				if err := os.Chdir(exeDir); err != nil {
					log.Printf("[Raven] Could not change directory to %s: %v", exeDir, err)
				}
			}
		}
	}

	config, err := server.LoadConfig("config.json")
	if err != nil {
		log.Printf("[Raven] Could not load config.json: %v", err)
	}

	storage := server.NewStorage(config.DataDir)
	if err := storage.Load(); err != nil {
		log.Fatalf("[Raven] Failed to load storage: %v", err)
	}

	store := storage.GetStore()
	changed, err := server.Migrate(store)
	if err != nil {
		log.Fatalf("[Raven] Migration failed: %v", err)
	}
	if changed {
		if err := storage.Save(); err != nil {
			log.Fatalf("[Raven] Failed to save post-migration: %v", err)
		}
	}

	srv := server.NewServer(config, storage)
	srv.SetupRoutes()

	log.Printf("[Raven] Starting Raven server on port %d...", config.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("[Raven] Server error: %v", err)
	}
}
