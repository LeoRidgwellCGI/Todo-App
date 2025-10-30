package service

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"todo-app/todo"
)

// TestService_ActorStore_Parallel validates concurrent access to the
// actor-backed Store. The actor should serialize writes and allow safe reads.
func TestService_ActorStore_Parallel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	st := NewActorStore(path)

	// allow actor to start & settle initial load
	time.Sleep(10 * time.Millisecond)

	// seed via Save through the actor
	if err := st.Save(ctx, []todo.Item{
		{ID: 1, Description: "seed", Status: todo.StatusNotStarted, CreatedAt: time.Now()},
	}); err != nil {
		t.Fatalf("seed Save: %v", err)
	}

	const writers = 4
	const readers = 12
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
			}
		}(r)
	}

	wg.Wait()
}
