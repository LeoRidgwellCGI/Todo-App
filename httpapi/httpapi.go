package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"todo-app/service"
	"todo-app/todo"
	"todo-app/trace"
)

// Register wires routes onto the provided mux using the given store.
func Register(mux *http.ServeMux, store service.Store) {
	// Add handler
	mux.HandleFunc("/add", withCtx(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var req struct {
			Description string `json:"description"`
			Status      string `json:"status"` // optional; default below
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(w, http.StatusBadRequest, err)
			return
		}

		desc := strings.TrimSpace(req.Description)
		if desc == "" {
			respondErr(w, http.StatusBadRequest, fmt.Errorf("description is required"))
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
			respondErr(w, http.StatusInternalServerError, err)
			return
		}

		// NOTE: todo.Add(list, description, status)
		list, item, err := todo.Add(list, desc, st)
		if err != nil {
			respondErr(w, http.StatusBadRequest, err)
			return
		}

		if err := store.Save(ctx, list); err != nil {
			respondErr(w, http.StatusInternalServerError, err)
			return
		}
		respondJSON(w, http.StatusCreated, item)
	}))

	// Get handler
	mux.HandleFunc("/get", withCtx(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// load list once
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(w, http.StatusInternalServerError, err)
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
		respondErr(w, http.StatusNotFound, fmt.Errorf("no to-do with id %d", id))
	}))

	// Update handler
	mux.HandleFunc("/update", withCtx(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(w, http.StatusBadRequest, err)
			return
		}
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(w, http.StatusInternalServerError, err)
			return
		}
		if req.Description != "" {
			list, err = todo.UpdateDescription(list, req.ID, strings.TrimSpace(req.Description))
			if err != nil {
				respondErr(w, http.StatusBadRequest, err)
				return
			}
		}

		if req.Status != "" {
			list, err = todo.UpdateStatus(list, req.ID, todo.Status(strings.TrimSpace(req.Status)))
			if err != nil {
				respondErr(w, http.StatusBadRequest, err)
				return
			}
		}

		if err := store.Save(ctx, list); err != nil {
			respondErr(w, http.StatusInternalServerError, err)
			return
		}

		if updated, ok := service.FindByID(list, req.ID); ok {
			respondJSON(w, http.StatusOK, updated)
			return
		}
	}))

	// Delete handler
	mux.HandleFunc("/delete", withCtx(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondErr(w, http.StatusBadRequest, err)
			return
		}
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(w, http.StatusInternalServerError, err)
			return
		}
		list, err = todo.Delete(list, req.ID)
		if err != nil {
			respondErr(w, http.StatusBadRequest, err)
			return
		}
		if err := store.Save(ctx, list); err != nil {
			respondErr(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	// Serve static /about/ from ./static/about
	mux.Handle("/about/", http.StripPrefix("/about/", http.FileServer(http.Dir("static/about"))))

	// (Optional) redirect /about -> /about/ so both work
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/about/", http.StatusMovedPermanently)
	})

	// Simple HTML view of all to-dos at /list
	mux.HandleFunc("/list", withCtx(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		list, err := store.Load(ctx)
		if err != nil {
			respondErr(w, http.StatusInternalServerError, err)
			return
		}
		tpl := template.Must(template.New("list").Parse(listTemplate))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, struct{ Items []todo.Item }{Items: list})
	}))
}

// withCtx injects a TraceID and passes context to a functional handler.
func withCtx(next func(context.Context, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if _, ok := trace.From(ctx); !ok {
			ctx, _ = trace.NewWithID(ctx, trace.GenerateID())
		}
		next(ctx, w, r.WithContext(ctx))
	}
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondErr(w http.ResponseWriter, status int, err error) {
	type errResp struct {
		Error string `json:"error"`
	}
	respondJSON(w, status, errResp{Error: err.Error()})
}

const listTemplate = "<!doctype html><html><head><meta charset=\"utf-8\"><title>Todos</title></head><body><h1>Todos</h1><ul>{{range .Items}}<li>{{.ID}} - {{.Description}} - {{.Status}}</li>{{else}}<li>none</li>{{end}}</ul></body></html>"
