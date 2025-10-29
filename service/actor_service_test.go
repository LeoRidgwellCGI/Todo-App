package service

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"todo-app/todo"
)

// TestActorStore_ConcurrentLoadAndSave verifies that ActorStore can handle
// concurrent Load and Save calls without data races or corruption.
func TestService_ActorStore_ConcurrentLoadAndSave(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	st := NewActorStore(path)
	defer st.Close()

	// write initial
	items := []todo.Item{
		{ID: 1, Description: "a", Status: "not started", CreatedAt: time.Now()},
	}
	if err := st.Save(ctx, items); err != nil {
		t.Fatalf("initial Save: %v", err)
	}

	// concurrent readers + single writer
	var wg sync.WaitGroup
	readers := 20
	wg.Add(readers)
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if _, err := st.Load(ctx); err != nil {
					t.Errorf("Load error: %v", err)
					return
				}
			}
		}()
	}

	// one writer updates the list
	go func() {
		time.Sleep(time.Millisecond / 2) // let some reads happen first (half a ms)
		items2 := append(items, todo.Item{ID: 2, Description: "b", Status: "started", CreatedAt: time.Now()})
		_ = st.Save(ctx, items2)
	}()

	wg.Wait()

	// verify file exists and can be loaded from disk via todo.Load
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file written: %v", err)
	}
	list, err := todo.Load(ctx, path)
	if err != nil {
		t.Fatalf("todo.Load: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}
