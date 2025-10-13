package trace

import (
	"context"
	"testing"
)

func TestGenerateIDLengthAndRandomness(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()
	if len(id1) != 32 || len(id2) != 32 {
		t.Fatalf("GenerateID length = %d,%d want 32", len(id1), len(id2))
	}
	if id1 == id2 {
		t.Fatalf("GenerateID produced duplicates")
	}
}

func TestNewAndFrom(t *testing.T) {
	base := context.Background()
	ctx, id := New(base)
	got, ok := From(ctx)
	if !ok || got == "" {
		t.Fatalf("From(ctx) missing id")
	}
	if got != id {
		t.Fatalf("From(ctx)=%q want=%q", got, id)
	}
}

func TestNewWithID(t *testing.T) {
	base := context.Background()
	ctx, id := NewWithID(base, "custom-id-123")
	if id != "custom-id-123" {
		t.Fatalf("NewWithID id=%q want custom-id-123", id)
	}
	got, ok := From(ctx)
	if !ok || got != "custom-id-123" {
		t.Fatalf("From(ctx)=%q ok=%v", got, ok)
	}

	// Empty should fallback to New()
	ctx2, id2 := NewWithID(base, "")
	if id2 == "" {
		t.Fatalf("NewWithID with empty id should generate one")
	}
	got2, ok := From(ctx2)
	if !ok || got2 != id2 {
		t.Fatalf("From(ctx2)=%q ok=%v want %q", got2, ok, id2)
	}
}
