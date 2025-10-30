package api_app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// --- parallel suite ---

// TestAPI_ParallelSuite spins up one server rooted in a temp working dir,
// then runs many subtests in parallel (readers/writers; HTML/static) to
// validate concurrency safety. Run with `go test -race` for best coverage.
func TestAPI_ParallelSuite(t *testing.T) {
	// Isolated working directory (single chdir done once; no t.Parallel here)
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	// Create ./static/about/index.html under the temp workspace for /about/
	if err := os.MkdirAll(filepath.Join("static", "about"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	const aboutHTML = "<!doctype html><html><body><h1>About OK</h1></body></html>"
	if err := os.WriteFile(filepath.Join("static", "about", "index.html"), []byte(aboutHTML), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	// Server writing to ./todos_test.json
	s := New("todos_test.json")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// Seed one item we will repeatedly read/update but never delete.
	var seed item
	{
		resp := doJSON(t, ts, "POST", "/add", map[string]any{
			"description": "seed",
			"status":      "not started",
		})
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed add status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}
		decodeJSON(t, resp.Body, &seed)
	}

	// Quick sanity that the output file exists before parallel chaos begins.
	if _, err := os.Stat(filepath.Join(tmp, "todos_test.json")); err != nil {
		t.Fatalf("expected %v/todos_test.json to exist: %v", tmp, err)
	}

	// ---------- Parallel subtests ----------
	// Readers hammer GET /get?id=<seed> repeatedly.
	for r := 0; r < 8; r++ {
		r := r
		t.Run(fmt.Sprintf("reader-%d", r), func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 50; i++ {
				resp := do(t, ts, "GET", "/get?id="+strconv.Itoa(seed.ID), nil)
				func() {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Fatalf("reader %d iter %d: status = %d", r, i, resp.StatusCode)
					}
					var got item
					decodeJSON(t, resp.Body, &got)
					if got.ID != seed.ID {
						t.Fatalf("reader %d iter %d: wrong id %d", r, i, got.ID)
					}
				}()
			}
		})
	}

	// Writers add new items continuously.
	for w := 0; w < 4; w++ {
		w := w
		t.Run(fmt.Sprintf("adder-%d", w), func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 40; i++ {
				desc := fmt.Sprintf("task-%d-%d", w, i)
				resp := doJSON(t, ts, "POST", "/add", map[string]any{
					"description": desc,
					"status":      "not started",
				})
				func() {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusCreated {
						t.Fatalf("adder %d iter %d: status = %d", w, i, resp.StatusCode)
					}
					var created item
					decodeJSON(t, resp.Body, &created)
					if created.Description != desc {
						t.Fatalf("adder %d iter %d: wrong description %q", w, i, created.Description)
					}
				}()
			}
		})
	}

	// Updaters repeatedly update the seed item (never delete it).
	for u := 0; u < 4; u++ {
		u := u
		t.Run(fmt.Sprintf("updater-%d", u), func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 40; i++ {
				status := "started"
				if i%2 == 1 {
					status = "completed"
				}
				desc := fmt.Sprintf("seed-upd-%d-%d", u, i)
				resp := doJSON(t, ts, "POST", "/update", map[string]any{
					"id":          seed.ID,
					"description": desc,
					"status":      status,
				})
				func() {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Fatalf("updater %d iter %d: status = %d", u, i, resp.StatusCode)
					}
				}()
			}
		})
	}

	// Creators + deleters: create a temp item then immediately delete it.
	for d := 0; d < 4; d++ {
		d := d
		t.Run(fmt.Sprintf("create-delete-%d", d), func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 30; i++ {
				resp := doJSON(t, ts, "POST", "/add", map[string]any{
					"description": fmt.Sprintf("temp-%d-%d", d, i),
					"status":      "not started",
				})
				var created item
				func() {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusCreated {
						t.Fatalf("c/d %d iter %d add: status = %d", d, i, resp.StatusCode)
					}
					decodeJSON(t, resp.Body, &created)
				}()
				resp2 := doJSON(t, ts, "POST", "/delete", map[string]any{"id": created.ID})
				func() {
					defer resp2.Body.Close()
					if resp2.StatusCode != http.StatusNoContent {
						t.Fatalf("c/d %d iter %d delete: status = %d", d, i, resp2.StatusCode)
					}
				}()
			}
		})
	}

	// HTML list endpoint should keep working while everything else is busy.
	for h := 0; h < 4; h++ {
		h := h
		t.Run(fmt.Sprintf("list-html-%d", h), func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 20; i++ {
				resp := do(t, ts, "GET", "/list", nil)
				func() {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Fatalf("list %d iter %d: status = %d", h, i, resp.StatusCode)
					}
					ct := resp.Header.Get("Content-Type")
					if !strings.Contains(ct, "text/html") {
						t.Fatalf("list %d iter %d: content-type = %q", h, i, ct)
					}
					// Read but don't assert specific items (they're changing); just ensure it's valid HTML-ish.
					_, _ = io.ReadAll(resp.Body)
				}()
			}
		})
	}

	// Static /about/ should also be fine under load.
	for a := 0; a < 2; a++ {
		a := a
		t.Run(fmt.Sprintf("about-%d", a), func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 10; i++ {
				resp := do(t, ts, "GET", "/about/", nil)
				func() {
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Fatalf("about %d iter %d: status = %d", a, i, resp.StatusCode)
					}
					ct := resp.Header.Get("Content-Type")
					if !strings.Contains(ct, "text/html") {
						t.Fatalf("about %d iter %d: content-type = %q", a, i, ct)
					}
				}()
			}
		})
	}

	// After all subtests, quick final consistency check.
	t.Run("final-consistency", func(t *testing.T) {
		// slight buffer for in-flight writes to complete
		time.Sleep(25 * time.Millisecond)
		resp := do(t, ts, "GET", "/get", nil)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("final get-all status = %d", resp.StatusCode)
		}
		var list []item
		decodeJSON(t, resp.Body, &list)
		if len(list) == 0 {
			t.Fatalf("final list unexpectedly empty")
		}
	})
}
