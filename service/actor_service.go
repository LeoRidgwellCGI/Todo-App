package service

import (
	"context"
	"log/slog"
	"time"

	"todo-app/todo"
)

// ActorStore is a concurrency-safe implementation of Store, using the
// actor pattern (a single goroutine owns the state and serializes writes).
// It allows many concurrent readers without locking the file and guarantees
// that writes are applied one-at-a-time.
//
// Zero shared mutable state is exposed; callers interact via messages.
type ActorStore struct {
	path string

	cmds chan any
	quit chan struct{}
}

// NewActorStore spins up the actor and loads the initial snapshot from disk.
// Use Close() to stop the background goroutine.
func NewActorStore(path string) *ActorStore {
	s := &ActorStore{
		path: path,
		cmds: make(chan any),
		quit: make(chan struct{}),
	}
	go s.loop()
	return s
}

// internal message types
type (
	getReq struct {
		ctx   context.Context
		reply chan []todo.Item
	}

	setReq struct {
		ctx   context.Context
		list  []todo.Item
		reply chan error
	}

	stopReq struct {
		done chan struct{}
	}
)

func (s *ActorStore) loop() {
	// private, goroutine-owned state
	var snapshot []todo.Item
	// load once at startup; treat missing file as empty list
	{
		ctx := context.Background()
		list, err := todo.Load(ctx, s.path)
		if err != nil {
			slog.Warn("actor: initial load failed; starting empty", "error", err, "path", s.path)
			list = []todo.Item{}
		}
		snapshot = cloneList(list)
	}

	for {
		select {
		case msg := <-s.cmds:
			switch m := msg.(type) {
			case getReq:
				// return a copy to avoid races with callers
				m.reply <- cloneList(snapshot)

			case setReq:
				// replace in-memory snapshot then persist to disk
				snapshot = cloneList(m.list)
				err := todo.Save(m.ctx, snapshot, s.path)
				m.reply <- err

			case stopReq:
				close(m.done)
				return
			}
		case <-s.quit:
			return
		}
	}
}

func cloneList(in []todo.Item) []todo.Item {
	out := make([]todo.Item, len(in))
	copy(out, in)
	return out
}

// Load returns a stable snapshot of the current list.
// Many callers can invoke Load concurrently without contention.
func (s *ActorStore) Load(ctx context.Context) ([]todo.Item, error) {
	reply := make(chan []todo.Item, 1)
	select {
	case s.cmds <- getReq{ctx: ctx, reply: reply}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case list := <-reply:
		return list, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Save sends a write to the actor and waits for it to complete.
// Writes are serialized; the actor also updates its in-memory snapshot.
func (s *ActorStore) Save(ctx context.Context, list []todo.Item) error {
	reply := make(chan error, 1)
	select {
	case s.cmds <- setReq{ctx: ctx, list: cloneList(list), reply: reply}:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close stops the actor gracefully.
func (s *ActorStore) Close() {
	done := make(chan struct{})
	select {
	case s.cmds <- stopReq{done: done}:
	case <-time.After(100 * time.Millisecond):
		// actor might be unresponsive; fall back to best-effort quit
	}
	select {
	case <-done:
	case <-time.After(150 * time.Millisecond):
	}
	close(s.quit)
}
