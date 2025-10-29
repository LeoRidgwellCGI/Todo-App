package service

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"todo-app/todo"
)

// TestService_EnsureOutPath verifies that FileStore.ensureOutPath
// returns the expected effective path based on various OutPath inputs.
func TestService_EnsureOutPath(t *testing.T) {
	t.Run("empty OutPath -> out/todos.json", func(t *testing.T) {
		f := &FileStore{OutPath: ""}
		got := f.ensureOutPath()
		want := filepath.Join("out", "todos.json")
		if got != want {
			t.Fatalf("ensureOutPath() = %q, want %q", got, want)
		}
	})

	t.Run("bare filename -> out/<filename>", func(t *testing.T) {
		f := &FileStore{OutPath: "mytodos.json"}
		got := f.ensureOutPath()
		want := filepath.Join("out", "mytodos.json")
		if got != want {
			t.Fatalf("ensureOutPath() = %q, want %q", got, want)
		}
	})

	t.Run("relative path with slash -> unchanged", func(t *testing.T) {
		f := &FileStore{OutPath: filepath.Join("subdir", "todos.json")}
		got := f.ensureOutPath()
		want := filepath.Join("subdir", "todos.json")
		if got != want {
			t.Fatalf("ensureOutPath() = %q, want %q", got, want)
		}
	})

	t.Run("absolute path -> unchanged", func(t *testing.T) {
		tmp := t.TempDir()
		abs := filepath.Join(tmp, "todos.json")
		f := &FileStore{OutPath: abs}
		got := f.ensureOutPath()
		if got != abs {
			t.Fatalf("ensureOutPath() = %q, want %q", got, abs)
		}
	})

	// Optional Windows separator check (only meaningful on windows)
	if runtime.GOOS == "windows" {
		t.Run("windows backslash in path -> unchanged", func(t *testing.T) {
			// Simulate "subdir\\todos.json" by joining then replacing on Windows.
			rel := filepath.Join("subdir", "todos.json")
			f := &FileStore{OutPath: rel}
			if f.ensureOutPath() != rel {
				t.Fatalf("ensureOutPath() changed a path with separators")
			}
		})
	}
}

// TestService_FileStore_SaveAndLoad_CreatesDirAndRoundTrips verifies that
// FileStore.Save creates necessary directories and that Load round-trips data.
// It uses a temporary directory for isolation.
// It verifies that the saved file exists at the expected path.
func TestService_FileStore_SaveAndLoad_CreatesDirAndRoundTrips(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	target := filepath.Join(tmp, "deep", "dir", "todos.json")

	f := &FileStore{OutPath: target}

	items := []todo.Item{
		{ID: 1, Description: "a", Status: "not started", CreatedAt: time.Unix(1700000000, 0).UTC()},
		{ID: 2, Description: "b", Status: "started", CreatedAt: time.Unix(1700000100, 0).UTC()},
	}
	if err := f.Save(ctx, items); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// File must exist at the nested path (directories created).
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected file to exist at %q, stat error = %v", target, err)
	}

	got, err := f.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got) != len(items) {
		t.Fatalf("Load() len = %d, want %d", len(got), len(items))
	}
	for i := range got {
		if got[i].ID != items[i].ID ||
			got[i].Description != items[i].Description ||
			got[i].Status != items[i].Status ||
			!got[i].CreatedAt.Equal(items[i].CreatedAt) {
			t.Fatalf("round-trip mismatch at %d: got %+v, want %+v", i, got[i], items[i])
		}
	}
}

// TestService_FileStore_SaveAndLoad_BareFilename_WritesUnderOutDir verifies that
// FileStore.Save and Load use the ./out/ directory when given a bare filename.
// It uses a temporary working directory for isolation.
// It verifies that the saved file exists at the expected path.
func TestService_FileStore_SaveAndLoad_BareFilename_WritesUnderOutDir(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	// Isolate the test by running inside a temp working directory.
	origWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}

	f := &FileStore{OutPath: "todos.json"}
	items := []todo.Item{
		{ID: 42, Description: "x", Status: "done", CreatedAt: time.Unix(1700000200, 0).UTC()},
	}
	if err := f.Save(ctx, items); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	wantPath := filepath.Join(tmp, "out", "todos.json")
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("expected bare filename to be saved under %q; stat error = %v", wantPath, err)
	}

	// Verify Load reads from the same effective path.
	got, err := f.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != 42 {
		t.Fatalf("Load() got %+v, want single item with ID 42", got)
	}
}

// TestService_FindByID verifies that FindByID correctly finds items by ID
// and returns not found when appropriate.
// It tests both found and not found cases.
func TestService_FindByID(t *testing.T) {
	list := []todo.Item{
		{ID: 1, Description: "a"},
		{ID: 2, Description: "b"},
	}
	item, ok := FindByID(list, 2)
	if !ok {
		t.Fatalf("FindByID(2) = not found, want found")
	}
	if item.ID != 2 || item.Description != "b" {
		t.Fatalf("FindByID(2) got %+v, want ID=2, Description=b", item)
	}

	_, ok = FindByID(list, 99)
	if ok {
		t.Fatalf("FindByID(99) = found, want not found")
	}
}
