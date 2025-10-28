package api_app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

// do sends a raw request (no automatic JSON encoding) and returns the response.
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
func decodeJSON(t *testing.T, r io.Reader, v any) {
	t.Helper()
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// TestAPI_Add_CreatesItem verifies that the /add endpoint creates a new to-do item
// and that the item is persisted to the expected file.
func TestAPI_Add_CreatesItem(t *testing.T) {
	// Isolated working directory
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	// Create server writing to ./out/todos_test.json
	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Create
	resp := doJSON(t, ts, "POST", "/add", map[string]any{
		"description": "Write tests",
		"status":      "not started",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var got item
	decodeJSON(t, resp.Body, &got)
	if got.Description != "Write tests" {
		t.Fatalf("description = %q", got.Description)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join("out", "todos_test.json")); err != nil {
		t.Fatalf("expected out/todos_test.json to exist: %v", err)
	}
}

// TestAPI_Get_ReturnsOne verifies that the /get endpoint returns one to-do item.
func TestAPI_Get_ReturnsOne(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Seed one item
	var created item
	{
		resp := doJSON(t, ts, "POST", "/add", map[string]any{"description": "Single task", "status": "not started"})
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("add status = %d", resp.StatusCode)
		}
		decodeJSON(t, resp.Body, &created)
	}

	// Get by ID
	resp := do(t, ts, "GET", "/get?id="+strconv.Itoa(created.ID), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get-by-id status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var got item
	decodeJSON(t, resp.Body, &got)
	if got.ID != created.ID || got.Description != created.Description {
		t.Fatalf("got item = %+v, want %+v", got, created)
	}
}

// TestAPI_Get_ReturnsAll verifies that the /get endpoint returns all to-do items.
func TestAPI_Get_ReturnsAll(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Seed two items
	_ = doJSON(t, ts, "POST", "/add", map[string]any{"description": "Alpha task", "status": "not started"})
	_ = doJSON(t, ts, "POST", "/add", map[string]any{"description": "Beta task", "status": "started"})

	// Get all
	resp := do(t, ts, "GET", "/get", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get-all status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var list []item
	decodeJSON(t, resp.Body, &list)
	if len(list) != 2 {
		t.Fatalf("list length = %d, want 2", len(list))
	}
}

// TestAPI_Update_ChangesFields verifies /update updates description and status.
func TestAPI_Update_ChangesFields(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Seed one item
	var created item
	{
		resp := doJSON(t, ts, "POST", "/add", map[string]any{"description": "Alpha", "status": "not started"})
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("add status = %d", resp.StatusCode)
		}
		decodeJSON(t, resp.Body, &created)
	}

	// UPDATE description and status
	updateBody := map[string]any{
		"id":          created.ID,
		"description": "Alpha updated",
		"status":      "completed",
	}
	resp := doJSON(t, ts, "POST", "/update", updateBody)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Verify via GET by id
	resp2 := do(t, ts, "GET", "/get?id="+strconv.Itoa(created.ID), nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("get-by-id status = %d, want %d", resp2.StatusCode, http.StatusOK)
	}
	var after item
	decodeJSON(t, resp2.Body, &after)
	if after.Description != "Alpha updated" || after.Status != "completed" {
		t.Fatalf("updated item = %+v", after)
	}
}

// TestAPI_Delete_RemovesItem verifies /delete removes the item and then list is empty.
func TestAPI_Delete_RemovesItem(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Seed one item
	var created item
	{
		resp := doJSON(t, ts, "POST", "/add", map[string]any{"description": "Alpha", "status": "started"})
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("add status = %d", resp.StatusCode)
		}
		decodeJSON(t, resp.Body, &created)
	}

	// DELETE
	resp := doJSON(t, ts, "POST", "/delete", map[string]any{"id": created.ID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	// Ensure now empty
	resp2 := do(t, ts, "GET", "/get", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("get-all status = %d, want %d", resp2.StatusCode, http.StatusOK)
	}
	var list2 []item
	decodeJSON(t, resp2.Body, &list2)
	if len(list2) != 0 {
		t.Fatalf("list length after delete = %d, want 0", len(list2))
	}
}

// TestAPI_ListRendersHTML validates the HTML table for /list
func TestAPI_ListRendersHTML(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Seed content
	_ = doJSON(t, ts, "POST", "/add", map[string]any{"description": "Alpha task", "status": "not started"})
	_ = doJSON(t, ts, "POST", "/add", map[string]any{"description": "Beta task", "status": "started"})

	resp := do(t, ts, "GET", "/list", nil)
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

// TestAPI_AboutServesStatic ensures /about/ serves content from ./static/about.
func TestAPI_AboutServesStatic(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	// Create ./static/about/index.html under the temp workspace
	if err := os.MkdirAll(filepath.Join("static", "about"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	const html = "<!doctype html><html><body><h1>About OK</h1></body></html>"
	if err := os.WriteFile(filepath.Join("static", "about", "index.html"), []byte(html), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp := do(t, ts, "GET", "/about/", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/about status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	ct := resp.Header.Get("Content-Type")
	// Allow either "text/html; charset=utf-8" or generic "text/html"
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("/about content-type = %q, want to contain %q", ct, "text/html")
	}
}
