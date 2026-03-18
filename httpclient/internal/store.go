package core

import "context"

// Store is the interface for persisting session data across client instances.
// Implementations must be safe for concurrent use.
type Store interface {
	Save(ctx context.Context, key string, value string) error
	Load(ctx context.Context) (map[string]string, error)
}
