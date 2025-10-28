package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"todo-app/service"
	"todo-app/todo"
	"todo-app/trace"
)

// CtxHandler defines a handler with context.
type CtxHandler func(context.Context, http.ResponseWriter, *http.Request)

// Register wires routes onto the provided mux using the given store.
func Register(mux *http.ServeMux, store service.Store) {
	// Handlers with logging and context injection
	mux.HandleFunc("/add", withCtx(logger(addHandler(store))))
	mux.HandleFunc("/get", withCtx(logger(getHandler(store))))
	mux.HandleFunc("/update", withCtx(logger(updateHandler(store))))
	mux.HandleFunc("/delete", withCtx(logger(deleteHandler(store))))
	mux.HandleFunc("/list", withCtx(logger(listHandler(store))))

	// Serve static /about/ from ./static/about
	mux.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("static/about"))))
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/about/", http.StatusMovedPermanently)
	})
}

// Add handler
func addHandler(store service.Store) CtxHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var req struct {
			Description string `json:"description"`
			Status      string `json:"status"` // optional; default below
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(ctx, w, http.StatusBadRequest, err)
			return
		}

		desc := strings.TrimSpace(req.Description)
		if desc == "" {
			respondErr(ctx, w, http.StatusBadRequest, fmt.Errorf("description is required"))
			return
		}

		// default status if none provided
		rawStatus := strings.TrimSpace(req.Status)
		if rawStatus == "" {
			rawStatus = "not started" // use whatever your app treats as the default
		}
		st := todo.Status(rawStatus)

		list, err := store.Load(ctx)
		if err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}

		// NOTE: todo.Add(list, description, status)
		list, item, err := todo.Add(list, desc, st)
		if err != nil {
			respondErr(ctx, w, http.StatusBadRequest, err)
			return
		}

		if err := store.Save(ctx, list); err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}
		respondJSON(w, http.StatusCreated, item)
	}
}

// Get handler
func getHandler(store service.Store) CtxHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// load list once
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}

		// if no id is provided -> return all
		idStr := strings.TrimSpace(r.URL.Query().Get("id"))
		if idStr == "" {
			respondJSON(w, http.StatusOK, list)
			return
		}

		// otherwise return single by id
		id, _ := strconv.Atoi(idStr)
		if it, ok := service.FindByID(list, id); ok {
			respondJSON(w, http.StatusOK, it)
			return
		}
		respondErr(ctx, w, http.StatusNotFound, fmt.Errorf("no to-do with id %d", id))
	}
}

// Update handler
func updateHandler(store service.Store) func(context.Context, http.ResponseWriter, *http.Request) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(ctx, w, http.StatusBadRequest, err)
			return
		}
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}
		if req.Description != "" {
			list, err = todo.UpdateDescription(list, req.ID, strings.TrimSpace(req.Description))
			if err != nil {
				respondErr(ctx, w, http.StatusBadRequest, err)
				return
			}
		}

		if req.Status != "" {
			list, err = todo.UpdateStatus(list, req.ID, todo.Status(strings.TrimSpace(req.Status)))
			if err != nil {
				respondErr(ctx, w, http.StatusBadRequest, err)
				return
			}
		}

		if err := store.Save(ctx, list); err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}

		if updated, ok := service.FindByID(list, req.ID); ok {
			respondJSON(w, http.StatusOK, updated)
			return
		}
	}
}

// Delete handler
func deleteHandler(store service.Store) CtxHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(ctx, w, http.StatusBadRequest, err)
			return
		}
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}
		list, err = todo.Delete(list, req.ID)
		if err != nil {
			respondErr(ctx, w, http.StatusBadRequest, err)
			return
		}
		if err := store.Save(ctx, list); err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// List handler - serves HTML page
func listHandler(store service.Store) CtxHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(ctx, w, http.StatusInternalServerError, err)
			return
		}
		tpl := template.Must(template.New("list").Parse(listTemplate))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, struct{ Items []todo.Item }{Items: list})
	}
}

// withCtx injects a TraceID and passes context to a functional handler.
func withCtx(next func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if _, ok := trace.From(ctx); !ok {
			ctx, _ = trace.NewWithID(ctx, trace.GenerateID())
			r = r.WithContext(ctx)
		}
		next(ctx, w, r)
	}
}

// statusRecorder captures status/bytes for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
func (s *statusRecorder) Write(b []byte) (int, error) {
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// logger emits start/end logs with trace_id, method, path, status and duration.
func logger(next CtxHandler) CtxHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		tid, _ := trace.From(ctx)
		start := time.Now()

		sr := &statusRecorder{ResponseWriter: w, status: 200}
		slog.InfoContext(ctx, "request start",
			"method", r.Method, "path", r.URL.Path, "trace_id", tid,
		)

		next(ctx, sr, r)

		dur := time.Since(start)
		fields := []any{
			"status", sr.status, "bytes", sr.bytes, "duration_ms", dur.Milliseconds(),
			"method", r.Method, "path", r.URL.Path, "trace_id", tid,
		}
		switch {
		case sr.status >= 500:
			slog.ErrorContext(ctx, "request end", fields...)
		case sr.status >= 400:
			slog.WarnContext(ctx, "request end", fields...)
		default:
			slog.InfoContext(ctx, "request end", fields...)
		}
	}
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// respondErr logs the error and responds with a JSON error message.
func respondErr(ctx context.Context, w http.ResponseWriter, status int, err error) {
	tid, _ := trace.From(ctx)
	slog.ErrorContext(ctx, "handler error", "status", status, "error", err, "trace_id", tid)
	type errResp struct {
		Error string `json:"error"`
	}
	respondJSON(w, status, errResp{Error: err.Error()})
}

const listTemplate = "<!doctype html><html><head><meta charset=\"utf-8\"><title>Todos</title></head><body><h1>Todos</h1><ul>{{range .Items}}<li>{{.ID}} - {{.Description}} - {{.Status}}</li>{{else}}<li>none</li>{{end}}</ul></body></html>"
