package core

import (
	"context"
	"net/http"
)

// CookieStore is the interface for persisting session data across client instances.
// Implementations must be safe for concurrent use.
type CookieStore interface {
	Save(ctx context.Context, cookies []*http.Cookie) error
	Load(ctx context.Context) ([]*http.Cookie, error)
}
