package todo

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// helper to make a small list for writing
func sampleItems(n int) []Item {
	out := make([]Item, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, Item{
			ID:          i,
			Description: "task",
			Status:      StatusNotStarted,
			CreatedAt:   time.Now(),
		})
	}
	return out
}

// TestTodo_SaveAndLoad_ParallelWritersReaders stresses Save and Load
// from many goroutines to the same file. The goal is to surface any
// partial-write corruption or racey reads. We don't assert exact list
// contents — only that operations succeed and JSON remains valid.
func TestTodo_SaveAndLoad_ParallelWritersReaders(t *testing.T) {
	t.Parallel() // allow this test to run alongside others

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	// seed the file once
	if err := Save(ctx, sampleItems(2), path); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	const writers = 8
	const readers = 8
	const iters = 50

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// writers: rewrite the file repeatedly
	for w := 0; w < writers; w++ {
		go func(wid int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				list := sampleItems((i % 3) + 1) // vary size 1..3
				if err := Save(ctx, list, path); err != nil {
					t.Errorf("writer %d Save(%d): %v", wid, i, err)
					return
				}
				// small pause to jitter interleavings
				time.Sleep(time.Duration((wid+i)%3) * time.Millisecond)
			}
		}(w)
	}

	// readers: continuously load
	for r := 0; r < readers; r++ {
		go func(rid int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				list, err := Load(ctx, path)
				if err != nil {
					t.Errorf("reader %d Load(%d): %v", rid, i, err)
					return
				}
				// lightweight sanity: list should never be nil and len bounded
				if list == nil {
					t.Errorf("reader %d got nil slice", rid)
					return
				}
				if len(list) > 10 {
					t.Errorf("reader %d unexpected large len=%d", rid, len(list))
					return
				}
			}
		}(r)
	}

	wg.Wait()
}
