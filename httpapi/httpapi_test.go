package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"todo-app/service"
	"todo-app/todo"
)

// --- test helpers & fakes ---

// memStore is a simple in-memory Store used for tests.
type memStore struct {
	list []todo.Item
}

func (m *memStore) Load(ctx context.Context) ([]todo.Item, error) {
	cp := make([]todo.Item, len(m.list))
	copy(cp, m.list)
	return cp, nil
}
func (m *memStore) Save(ctx context.Context, list []todo.Item) error {
	cp := make([]todo.Item, len(list))
	copy(cp, list)
	m.list = cp
	return nil
}
func (m *memStore) seed(items []todo.Item) { m.list = append([]todo.Item(nil), items...) }

// decodeJSON reads the response body and JSON-decodes into v.
func decodeJSON(t *testing.T, r *http.Response, v any) {
	t.Helper()
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("decode json: %v; body=%s", err, string(data))
	}
}

// newMuxWithStore creates a new http.ServeMux with the given Store registered.
func newMuxWithStore(s service.Store) *http.ServeMux {
	// Silence logs during tests.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
	m := http.NewServeMux()
	Register(m, s)
	return m
}

// --- tests ---

// TestHTTPAPI_AddGetUpdateDelete_Flow tests the full flow of adding, getting, updating, and deleting a to-do item
// via the HTTP API handlers.
func TestHTTPAPI_AddGetUpdateDelete_Flow(t *testing.T) {
	store := &memStore{}
	mux := newMuxWithStore(store)

	// 1) Add
	addReq := map[string]any{"description": "write tests"}
	body, _ := json.Marshal(addReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("add status=%d, want %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	var created todo.Item
	decodeJSON(t, w.Result(), &created)
	if strings.TrimSpace(created.Description) != "write tests" {
		t.Fatalf("created.Description=%q, want %q", created.Description, "write tests")
	}
	if strings.TrimSpace(string(created.Status)) == "" {
		t.Fatalf("created.Status should be set (got empty)")
	}

	// 2) Get list (no id)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get(all) status=%d, want %d", w.Code, http.StatusOK)
	}
	var gotList []todo.Item
	decodeJSON(t, w.Result(), &gotList)
	if len(gotList) != 1 {
		t.Fatalf("get(all) len=%d, want 1", len(gotList))
	}

	id := gotList[0].ID

	// 3) Get by id
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get?id="+itoa(id), nil)
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get(id) status=%d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var gotItem todo.Item
	decodeJSON(t, w.Result(), &gotItem)
	if gotItem.ID != id {
		t.Fatalf("get(id).ID=%d, want %d", gotItem.ID, id)
	}

	// 4) Update description + status
	upReq := map[string]any{
		"id":          id,
		"description": "write more tests",
		"status":      "started",
	}
	body, _ = json.Marshal(upReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update status=%d, want %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var updated todo.Item
	decodeJSON(t, w.Result(), &updated)
	if updated.Description != "write more tests" || string(updated.Status) != "started" {
		t.Fatalf("updated item mismatch: %+v", updated)
	}

	// 5) Delete
	delReq := map[string]any{"id": id}
	body, _ = json.Marshal(delReq)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d, want %d; body=%s", w.Code, http.StatusNoContent, w.Body.String())
	}

	// Check it is gone via /get?id=
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get?id="+itoa(id), nil)
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get(after delete) status=%d, want %d; body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

// TestHTTPAPI_Add_BadRequest_WhenNoDescription verifies that the Add handler
// returns a 400 Bad Request when given an empty description.
func TestHTTPAPI_Add_BadRequest_WhenNoDescription(t *testing.T) {
	store := &memStore{}
	mux := newMuxWithStore(store)

	body, _ := json.Marshal(map[string]any{"description": "   "})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	var er struct{ Error string }
	decodeJSON(t, w.Result(), &er)
	if er.Error == "" {
		t.Fatalf("expected error message body")
	}
}

// TestHTTPAPI_Get_All_WhenEmpty verifies that the Get handler
// returns an empty list when there are no to-do items.
func TestHTTPAPI_Get_All_WhenEmpty(t *testing.T) {
	store := &memStore{}
	mux := newMuxWithStore(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d", w.Code, http.StatusOK)
	}
	var list []todo.Item
	decodeJSON(t, w.Result(), &list)
	if len(list) != 0 {
		t.Fatalf("len(list)=%d, want 0", len(list))
	}
}

// TestHTTPAPI_List_HTML_Render verifies that the /list handler
// correctly renders an HTML page with to-do items.
func TestHTTPAPI_List_HTML_Render(t *testing.T) {
	// Seed store with one item so the template renders a list entry.
	store := &memStore{}
	store.seed([]todo.Item{
		{
			ID:          7,
			Description: "render me",
			Status:      "not started",
			CreatedAt:   time.Now(),
		},
	})
	mux := newMuxWithStore(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	mux.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type=%q, want text/html", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "render me") || !strings.Contains(body, "not started") {
		t.Fatalf("HTML body missing expected content: %q", body)
	}
}

// TestHTTPAPI_About_StaticRedirectAndFiles tests the /about redirect and static file serving.
// It verifies that /about redirects to /about/ and that /about/ serves static files.
func TestHTTPAPI_About_StaticRedirectAndFiles(t *testing.T) {
	// This test ensures the /about redirect and static handler are registered.
	store := &memStore{}
	mux := newMuxWithStore(store)

	// 1) Redirect /about -> /about/
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusMovedPermanently {
		t.Fatalf("/about redirect status=%d, want %d", w.Code, http.StatusMovedPermanently)
	}
	loc := w.Header().Get("Location")
	if loc != "/about/" {
		t.Fatalf("redirect Location=%q, want /about/", loc)
	}

	// 2) /about/ is served from static/about; we don't require real files,
	// but ensure the handler is wired and returns 404 (not 500) for a missing file.
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/about/", nil)
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound && w.Code != http.StatusOK {
		t.Fatalf("/about/ should be served by FileServer (status=%d)", w.Code)
	}

	// Optional: demonstrate that a real file would be served if present.
	// (Create a temp file under static/about and request it.)
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	_ = osWriteFile(indexPath, []byte("<h1>About OK</h1>"))

	// Re-register mux with a custom dir by temporarily chdir.
	origWD, _ := osGetwd()
	t.Cleanup(func() { _ = osChdir(origWD) })
	_ = osChdir(dir)
	mux = http.NewServeMux()
	// Re-register with static pointing at ./static/about (we'll create that structure)
	_ = osMkdirAll(filepath.Join("static", "about"), 0o755)
	_ = osRename("index.html", filepath.Join("static", "about", "index.html"))
	Register(mux, store)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/about/", nil)
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "About OK") {
		t.Fatalf("expected static about page served; status=%d body=%q", w.Code, w.Body.String())
	}
}

// itoa is a tiny helper to avoid importing strconv in tests.
func itoa(i int) string { return strconvItoa(i) }

// --- tiny stdlib shims (keep imports tidy in this file) ---

// The following wrappers keep imports minimal & obvious above.
func osWriteFile(name string, data []byte) error { return os.WriteFile(name, data, 0o644) }
func osGetwd() (string, error)                   { return os.Getwd() }
func osChdir(dir string) error                   { return os.Chdir(dir) }
func osMkdirAll(path string, perm uint32) error  { return os.MkdirAll(path, os.FileMode(perm)) }
func osRename(old, new string) error             { return os.Rename(old, new) }

// re-export strconv.Itoa without importing it at top (to keep the import list short above)
func strconvItoa(i int) string {
	return func(x int) string {
		const digits = "0123456789"
		if x == 0 {
			return "0"
		}
		neg := x < 0
		if neg {
			x = -x
		}
		var buf [32]byte
		n := len(buf)
		for x > 0 {
			n--
			buf[n] = digits[x%10]
			x /= 10
		}
		if neg {
			n--
			buf[n] = '-'
		}
		return string(buf[n:])
	}(i)
}
