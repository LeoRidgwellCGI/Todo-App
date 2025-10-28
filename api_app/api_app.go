// api_app/api_app.go
// Package api_app implements the Todo API server.
package api_app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"todo-app/httpapi"
	"todo-app/service"
)

// Server is now a thin bootstrapper (intentionally small).
// All HTTP concerns (routing + handlers) live in package httpapi.
type Server struct {
	store service.Store
	mux   *http.ServeMux
}

// New constructs a server using a JSON file at outPath.
func New(outPath string) *Server {
	st := service.NewFileStore(outPath)
	mux := http.NewServeMux()
	httpapi.Register(mux, st)
	return &Server{store: st, mux: mux}
}

// Handler returns the fully wired HTTP handler.
func (s *Server) Handler() http.Handler { return s.mux }

// Run starts the HTTP server at addr and shuts down on ctx.Done().
func (s *Server) Run(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: s.mux}
	go func() {
		<-ctx.Done()
		slog.Info("shutting down server")
		_ = srv.Shutdown(context.Background())
	}()
	slog.Info("listening", "addr", addr)
	return srv.ListenAndServe()
}

// FromEnv constructs a Server and derives the address from PORT, like Heroku.
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
