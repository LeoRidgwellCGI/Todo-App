package todo

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

//
// todo/storage.go (package todo)
// ------------------------------
// JSON persistence helpers. These are the *only* functions that touch disk.
// They are context-aware so logs include the trace_id set at process start.
//

// ensureParentDir ensures the directory for the provided file path exists.
// It is safe to call even if the directory already exists.
func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o755)
}

// Save serializes the given list to pretty-printed JSON and writes to `path`.
// It ensures the parent directory exists (e.g., ./out/). On success, an info log
// is emitted containing the path and the number of items.
func Save(ctx context.Context, list []Item, path string) error {
	// 1) Ensure ./out/ exists (or any parent directory for the provided path).
	if err := ensureParentDir(path); err != nil {
		slog.ErrorContext(ctx, "failed to create output directory", "error", err, "path", path)
		return err
	}

	// 2) Marshal to JSON (readable formatting to make diffs easier in VCS).
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal todos", "error", err, "path", path)
		return err
	}

	// 3) Write with owner-readable defaults.
	if err := os.WriteFile(path, data, 0o644); err != nil {
		slog.ErrorContext(ctx, "failed to save todos", "error", err, "path", path)
		return err
	}

	// 4) Log success with structured attributes for observability.
	slog.InfoContext(ctx, "todos saved", "path", path, "count", len(list))
	return nil
}

// Load reads a JSON file at `path`. If the file does not exist, we return an empty list.
// Any parse or read error is logged and returned to the caller.
func Load(ctx context.Context, path string) ([]Item, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		// Missing file is not an error — callers expect an empty list initially.
		if errors.Is(err, fs.ErrNotExist) {
			return []Item{}, nil
		}
		slog.ErrorContext(ctx, "failed to read file", "error", err, "path", path)
		return nil, err
	}
	if len(b) == 0 {
		return []Item{}, nil
	}

	var list []Item
	if err := json.Unmarshal(b, &list); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal JSON", "error", err, "path", path)
		return nil, err
	}
	return list, nil
}
