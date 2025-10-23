// api_app/api_app.go
// Package api_app implements the Todo API server.
package api_app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"todo-app/todo"
	"todo-app/trace"
)

// Server exposes HTTP endpoints for the Todo app.
type Server struct {
	outPath string
	mux     *http.ServeMux
}

// New constructs a Server with routes registered on a ServeMux.
func New(outPath string) *Server {
	s := &Server{
		outPath: outPath,
		mux:     http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler returns the http.Handler for this API server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// Register routes on the internal ServeMux.
// Endpoints: /get, /add, /update, /delete
func (s *Server) registerRoutes() {
	// API endpoints
	s.mux.HandleFunc("/get", s.withCtx(s.handleGet))
	s.mux.HandleFunc("/add", s.withCtx(s.handleAdd))
	s.mux.HandleFunc("/update", s.withCtx(s.handleUpdate))
	s.mux.HandleFunc("/delete", s.withCtx(s.handleDelete))

	// Serve static "about" page
	fs := http.FileServer(http.Dir("static"))
	s.mux.Handle("/about/", http.StripPrefix("/about/", fs))

	// Health
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}

// withCtx is a helper to wrap handlers with context that includes TraceID.
// It checks for an incoming X-Trace-ID header to propagate.
func (s *Server) withCtx(next func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if h := r.Header.Get("X-Trace-ID"); h != "" {
			ctx = func() context.Context { ctx2, _ := trace.NewWithID(ctx, h); return ctx2 }()
		} else {
			ctx = func() context.Context { ctx2, _ := trace.New(ctx); return ctx2 }()
		}

		next(ctx, w, r.WithContext(ctx))
	}
}

// respondJSON writes v as JSON with the given status code.
// Ignores encoding errors.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// respondErr logs the error and responds with a JSON error message.
// Ignores encoding errors.
func respondErr(ctx context.Context, w http.ResponseWriter, status int, err error) {
	slog.ErrorContext(ctx, "request failed", "status", status, "error", err)
	type errResp struct {
		Error string `json:"error"`
	}
	respondJSON(w, status, errResp{Error: err.Error()})
}

// Data access helpers
// ensureOutPath returns the normalized output path for the todo JSON file.
func (s *Server) ensureOutPath() string {
	p := s.outPath
	if p == "" {
		p = "out/todos.json"
	}
	// normalize to ./out/<basename>
	base := filepath.Base(p)
	return filepath.ToSlash(filepath.Join("out", base))
}

// load reads the todo items from the output path.
func (s *Server) load(ctx context.Context) ([]todo.Item, error) {
	return todo.Load(ctx, s.ensureOutPath())
}

// save writes the todo items to the output path.
func (s *Server) save(ctx context.Context, list []todo.Item) error {
	return todo.Save(ctx, list, s.ensureOutPath())
}

// findByID searches for an item by ID in the list.
// Returns (item, true) if found; (zero, false) if not found.
func (s *Server) findByID(list []todo.Item, id int) (todo.Item, bool) {
	for _, it := range list {
		if it.ID == id {
			return it, true
		}
	}
	return todo.Item{}, false
}

// Handlers
// GET /get or /get?id=123
// If id is provided, returns that item; else returns full list.
func (s *Server) handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondErr(ctx, w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	list, err := s.load(ctx)
	if err != nil {
		respondErr(ctx, w, http.StatusInternalServerError, err)
		return
	}

	idStr := r.URL.Query().Get("id")
	if strings.TrimSpace(idStr) == "" {
		respondJSON(w, http.StatusOK, list)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErr(ctx, w, http.StatusBadRequest, fmt.Errorf("invalid id %q: %w", idStr, err))
		return
	}

	if it, ok := s.findByID(list, id); ok {
		respondJSON(w, http.StatusOK, it)
		return
	}

	respondErr(ctx, w, http.StatusNotFound, fmt.Errorf("no to-do with id %d", id))
}

// POST /add  body: {"description":"...", "status":"not started|started|completed"}
type addReq struct {
	Description string      `json:"description"`
	Status      todo.Status `json:"status"`
}

// handleAdd handles creating a new todo item.
// Expects a POST with JSON body.
func (s *Server) handleAdd(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondErr(ctx, w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req addReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(ctx, w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	list, err := s.load(ctx)
	if err != nil {
		respondErr(ctx, w, http.StatusInternalServerError, err)
		return
	}

	list, item, err := todo.Add(list, req.Description, req.Status)
	if err != nil {
		respondErr(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := s.save(ctx, list); err != nil {
		respondErr(ctx, w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

// POST|PATCH /update  body: {"id":123, "description":"...", "status":"..."}
type updateReq struct {
	ID          int          `json:"id"`
	Description *string      `json:"description,omitempty"`
	Status      *todo.Status `json:"status,omitempty"`
}

// handleUpdate handles updating an existing todo item.
// Expects a POST or PATCH with JSON body.
func (s *Server) handleUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		respondErr(ctx, w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(ctx, w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if req.ID == 0 {
		respondErr(ctx, w, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	list, err := s.load(ctx)
	if err != nil {
		respondErr(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if req.Description != nil {
		if list, err = todo.UpdateDescription(list, req.ID, *req.Description); err != nil {
			respondErr(ctx, w, http.StatusBadRequest, err)
			return
		}
	}

	if req.Status != nil {
		if list, err = todo.UpdateStatus(list, req.ID, *req.Status); err != nil {
			respondErr(ctx, w, http.StatusBadRequest, err)
			return
		}
	}

	if err := s.save(ctx, list); err != nil {
		respondErr(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if it, ok := s.findByID(list, req.ID); ok {
		respondJSON(w, http.StatusOK, it)
		return
	}

	respondErr(ctx, w, http.StatusNotFound, fmt.Errorf("no to-do with id %d", req.ID))
}

// POST|DELETE /delete  body: {"id":123}
type deleteReq struct {
	ID int `json:"id"`
}

// handleDelete handles deleting a todo item.
// Expects a POST or DELETE with JSON body.
func (s *Server) handleDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		respondErr(ctx, w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	// Parse request
	var req deleteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(ctx, w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if req.ID == 0 {
		respondErr(ctx, w, http.StatusBadRequest, errors.New("id is required"))
		return
	}

	list, err := s.load(ctx)
	// Data operations
	if err != nil {
		respondErr(ctx, w, http.StatusInternalServerError, err)
		return
	}

	newList, err := todo.Delete(list, req.ID)
	if err != nil {
		respondErr(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := s.save(ctx, newList); err != nil {
		respondErr(ctx, w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": req.ID})
}

// Run listens and serves on addr until context cancel.
func (s *Server) Run(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	// Wait for context done or server error.
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

// Convenience to construct with env vars.
func FromEnv() (*Server, string) {
	addr := ":8080"
	if v := os.Getenv("PORT"); strings.TrimSpace(v) != "" {
		addr = ":" + strings.TrimPrefix(v, ":")
	}
	outPath := "out/todos.json"
	if v := os.Getenv("TODO_OUT"); strings.TrimSpace(v) != "" {
		outPath = v
	}
	return New(outPath), addr
}
