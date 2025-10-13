package trace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

//
// trace/trace.go (package trace)
// ------------------------------
// Minimal trace helper to generate and propagate a TraceID through context.
// We keep the key private to prevent collisions with other context users.
//

// keyType is an unexported type to prevent collisions in context values.
type keyType struct{}

// key is the package-private context key for the trace id value.
var key keyType

// New creates a new context with a generated TraceID and returns (ctx, id).
func New(parent context.Context) (context.Context, string) {
	id := GenerateID()
	return context.WithValue(parent, key, id), id
}

// NewWithID places the provided id into the context, or generates one if empty.
// It returns the derived context and the id used.
func NewWithID(parent context.Context, id string) (context.Context, string) {
	if id == "" {
		return New(parent)
	}
	return context.WithValue(parent, key, id), id
}

// From retrieves the TraceID stored in ctx, if any.
// Returns (id, true) on success; ("", false) if missing.
func From(ctx context.Context) (string, bool) {
	v := ctx.Value(key)
	if s, ok := v.(string); ok && s != "" {
		return s, true
	}
	return "", false
}

// GenerateID returns a random 16-byte hex string (32 hex chars).
// We use crypto/rand for strong uniqueness properties.
func GenerateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Unlikely path — return zeros if entropy source fails.
		return hex.EncodeToString(b[:])
	}
	return hex.EncodeToString(b[:])
}
