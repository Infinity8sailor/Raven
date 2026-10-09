package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	storage   *Storage
	config    *Config
	startTime time.Time
	mux       *http.ServeMux
}

func NewServer(config *Config, storage *Storage) *Server {
	return &Server{
		storage:   storage,
		config:    config,
		startTime: time.Now(),
		mux:       http.NewServeMux(),
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, msg, details string) {
	writeJSON(w, status, APIError{Error: msg, Details: details})
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *Server) SetupRoutes() {
	apiMux := http.NewServeMux()
	
	apiMux.HandleFunc("GET /lists", s.handleGetLists)
	apiMux.HandleFunc("POST /lists", s.handleCreateList)
	apiMux.HandleFunc("PUT /lists/{id}", s.handleUpdateList)
	apiMux.HandleFunc("DELETE /lists/{id}", s.handleDeleteList)
	apiMux.HandleFunc("PATCH /lists/{id}/features", s.handleUpdateFeatures)
	apiMux.HandleFunc("POST /lists/reorder", s.handleReorderLists)

	apiMux.HandleFunc("POST /lists/{id}/items", s.handleCreateItem)
	apiMux.HandleFunc("PUT /lists/{id}/items/{itemId}", s.handleUpdateItem)
	apiMux.HandleFunc("DELETE /lists/{id}/items/{itemId}", s.handleDeleteItem)
	apiMux.HandleFunc("PATCH /lists/{id}/items/{itemId}/status", s.handleToggleStatus)
	apiMux.HandleFunc("POST /lists/{id}/items/reorder", s.handleReorderItems)

	apiMux.HandleFunc("GET /search", s.handleSearch)
	apiMux.HandleFunc("GET /settings", s.handleGetSettings)
	apiMux.HandleFunc("PUT /settings", s.handleUpdateSettings)
	apiMux.HandleFunc("POST /backup", s.handleBackup)
	apiMux.HandleFunc("GET /health", s.handleHealth)

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		apiMux.ServeHTTP(w, r)
	})

	s.mux.Handle("/api/", http.StripPrefix("/api", apiHandler))
	s.mux.Handle("/", http.FileServer(http.Dir(s.config.WebDir)))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	return http.ListenAndServe(addr, s)
}

func findList(store *TaskStore, id string) (*List, int) {
	for i := range store.Lists {
		if store.Lists[i].ID == id {
			return &store.Lists[i], i
		}
	}
	return nil, -1
}

func findItem(list *List, id string) (*Item, int) {
	for i := range list.Items {
		if list.Items[i].ID == id {
			return &list.Items[i], i
		}
	}
	return nil, -1
}
