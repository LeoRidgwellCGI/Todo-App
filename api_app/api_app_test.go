package api_app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type item struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// --- test helpers ---

// doJSON sends a JSON request to the test server and returns the response.
// Fails the test on error.
func doJSON(t *testing.T, ts *httptest.Server, method, path string, payload any) *http.Response {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, ts.URL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

// do sends a request to the test server and returns the response.
// Fails the test on error.
func do(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, ts.URL+path, body)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

// decodeJSON decodes JSON from r into v. Fails the test on error.
// v must be a pointer.
func decodeJSON(t *testing.T, r io.Reader, v any) {
	t.Helper()
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// TestAPI_AddGetUpdateDelete exercises the API happy path
// in a temporary working directory so that "./out" is sandboxed per test run.
// The test covers adding an item, retrieving it, updating its description and status, and deleting it.
// At each step, the test verifies the expected outcomes.
// The server is constructed to write to "out/todos_test.json" under the temp dir.
// The test uses the httptest package to create a test server.
func TestAPI_AddGetUpdateDelete(t *testing.T) {
	// Create an isolated working directory
	tmp := t.TempDir()

	// Save original CWD and ensure we restore it before TempDir cleanup.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	// Run *before* the TempDir is removed (Cleanup runs LIFO; TempDir’s cleanup was registered earlier)
	t.Cleanup(func() {
		_ = os.Chdir(cwd) // best-effort restore
	})

	// Construct server that writes to ./out/<basename> under the temp dir.
	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// ---- ADD ----
	addBody := map[string]any{
		"description": "Write tests",
		"status":      "started",
	}
	var created item
	{
		resp := doJSON(t, ts, "POST", "/add", addBody)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}
		decodeJSON(t, resp.Body, &created)
		if created.ID <= 0 {
			t.Fatalf("expected positive ID, got %d", created.ID)
		}
		if created.Description != "Write tests" {
			t.Fatalf("desc = %q, want %q", created.Description, "Write tests")
		}
		if created.Status != "started" {
			t.Fatalf("status = %q, want %q", created.Status, "started")
		}
		if _, err := os.Stat(filepath.FromSlash("out/todos_test.json")); err != nil {
			t.Fatalf("expected out/todos_test.json to exist: %v", err)
		}
	}

	// ---- GET ALL ----
	var list []item
	{
		resp := do(t, ts, "GET", "/get", nil)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("get-all status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		decodeJSON(t, resp.Body, &list)
		if len(list) != 1 {
			t.Fatalf("list length = %d, want %d", len(list), 1)
		}
	}

	// ---- GET ONE ----
	{
		u := ts.URL + "/get?id=" + url.QueryEscape(strconv.Itoa(created.ID))
		resp, err := http.Get(u)
		if err != nil {
			t.Fatalf("GET one: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("get-one status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		var got item
		decodeJSON(t, resp.Body, &got)
		if got.ID != created.ID {
			t.Fatalf("get-one id = %d, want %d", got.ID, created.ID)
		}
	}

	// ---- UPDATE description ----
	{
		upd := map[string]any{
			"id":          created.ID,
			"description": "Write more tests",
		}
		resp := doJSON(t, ts, "POST", "/update", upd)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("update desc status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		var got item
		decodeJSON(t, resp.Body, &got)
		if got.Description != "Write more tests" {
			t.Fatalf("updated desc = %q, want %q", got.Description, "Write more tests")
		}
	}

	// ---- UPDATE status ----
	{
		upd := map[string]any{
			"id":     created.ID,
			"status": "completed",
		}
		resp := doJSON(t, ts, "POST", "/update", upd)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("update status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		var got item
		decodeJSON(t, resp.Body, &got)
		if got.Status != "completed" {
			t.Fatalf("updated status = %q, want %q", got.Status, "completed")
		}
	}

	// ---- DELETE ----
	{
		del := map[string]any{"id": created.ID}
		resp := doJSON(t, ts, "POST", "/delete", del)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("delete status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		// Now list should be empty
		resp2 := do(t, ts, "GET", "/get", nil)
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("get-all(2) status = %d, want %d", resp2.StatusCode, http.StatusOK)
		}
		var list2 []item
		decodeJSON(t, resp2.Body, &list2)
		if len(list2) != 0 {
			t.Fatalf("list length after delete = %d, want %d", len(list2), 0)
		}
	}
}

// Verifies the static /about/ endpoint (served with http.FileServer)
// returns 200 OK and serves HTML content. The test creates a temporary
// ./static/about/index.html under the per-test working directory.
func TestAPI_AboutServesStatic(t *testing.T) {
	// Create an isolated working directory
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	// Arrange a static file under ./static/about/index.html
	if err := os.MkdirAll("static/about", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	const html = "<!doctype html><html><head><title>About</title></head><body>About Todo-App</body></html>"
	if err := os.WriteFile("static/about/index.html", []byte(html), 0o644); err != nil {
		t.Fatalf("write about index: %v", err)
	}

	// Start the API server (it should have /about/ mounted via http.FileServer)
	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Request the about page (directory path with trailing slash so index.html is served)
	resp := do(t, ts, http.MethodGet, "/about/", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("about status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("about content-type = %q, want to contain %q", ct, "text/html")
	}
}

// TestAPI_ListRendersHTML verifies the dynamic /list page (served via html/template)
// renders a table of todos and includes the descriptions of created items.
func TestAPI_ListRendersHTML(t *testing.T) {
	// Create an isolated working directory
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	// Start server
	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Add a couple of todos (use the same path style the file already uses)
	resp1 := doJSON(t, ts, http.MethodPost, "/add", map[string]any{
		"description": "Alpha task",
		"status":      "started",
	})
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("create(1) status = %d, want %d", resp1.StatusCode, http.StatusCreated)
	}

	// Add another todo
	resp2 := doJSON(t, ts, http.MethodPost, "/add", map[string]any{
		"description": "Beta task",
		"status":      "not started",
	})
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusCreated {
		t.Fatalf("create(2) status = %d, want %d", resp2.StatusCode, http.StatusCreated)
	}

	// Request /list (HTML view)
	resp := do(t, ts, http.MethodGet, "/list", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/list status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("/list content-type = %q, want to contain %q", ct, "text/html")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	html := string(body)
	if !strings.Contains(html, "Alpha task") || !strings.Contains(html, "Beta task") {
		t.Fatalf("/list HTML did not include expected items; got:\n%s", html)
	}
}
