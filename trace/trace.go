package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type keyType struct{}

var key keyType

// New creates a new context with a generated TraceID and returns (ctx, id).
func New(parent context.Context) (context.Context, string) {
	id := GenerateID()
	ctx := context.WithValue(parent, key, id)
	return ctx, id
}

// NewWithID stores the provided id into the context; if empty, generates one.
// Returns (ctx, idUsed).
func NewWithID(parent context.Context, id string) (context.Context, string) {
	if id == "" {
		return New(parent)
	}
	ctx := context.WithValue(parent, key, id)
	return ctx, id
}

// From returns the TraceID stored in the context, if any.
func From(ctx context.Context) (string, bool) {
	v := ctx.Value(key)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// GenerateID returns a random 16-byte hex string (32 hex chars).
func GenerateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failure is extremely unlikely; fall back to zeros.
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString(b[:])
}
