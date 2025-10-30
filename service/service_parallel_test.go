package service

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"todo-app/todo"
)

// TestService_FileStore_ParallelLoadersAndWriters validates that FileStore is
// safe under concurrent Load/Save. Use `go test -race` to enable the data race
// detector in addition to these assertions.
func TestService_FileStore_ParallelLoadersAndWriters(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	st := &FileStore{OutPath: path}

	// seed file
	if err := st.Save(ctx, []todo.Item{
		{ID: 1, Description: "a", Status: todo.StatusNotStarted, CreatedAt: time.Now()},
	}); err != nil {
		t.Fatalf("seed Save: %v", err)
	}

	const writers = 6
	const readers = 10
	const iters = 40

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for w := 0; w < writers; w++ {
		go func(wid int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				list := []todo.Item{
					{ID: 1, Description: "a", Status: todo.StatusNotStarted, CreatedAt: time.Now()},
					{ID: 2, Description: "b", Status: todo.StatusNotStarted, CreatedAt: time.Now()},
				}
				if err := st.Save(ctx, list); err != nil {
					t.Errorf("writer %d Save(%d): %v", wid, i, err)
					return
				}
				time.Sleep(time.Duration((wid+i)%3) * time.Millisecond)
			}
		}(w)
	}

	for r := 0; r < readers; r++ {
		go func(rid int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				list, err := st.Load(ctx)
				if err != nil {
					t.Errorf("reader %d Load(%d): %v", rid, i, err)
					return
				}
				if list == nil {
					t.Errorf("reader %d got nil slice", rid)
					return
				}
				// no expectations on exact length; just ensure sane upper bound
				if len(list) > 10 {
					t.Errorf("reader %d got suspicious length=%d", rid, len(list))
					return
				}
			}
		}(r)
	}

	wg.Wait()
}
