package app

import (
	"fmt"
	"os"
	"path/filepath"

	httpclient "github.com/sengokyu/kusabase/httpclient"

	"github.com/sengokyu/kusabase/cli/internal/session"
	"github.com/sengokyu/kusabase/cli/internal/storage"
)

const sessionCookieKey = "next-session"

// App holds the HTTP client and session state for a CLI invocation.
type App struct {
	Client  *httpclient.Client
	Session *session.Session
	store   *storage.FileStore
}

// New creates an App from environment configuration.
// KUSA_BASE_URL must be set.
func New() (*App, error) {
	baseURL := os.Getenv("KUSA_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("KUSA_BASE_URL is not set")
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine cache directory: %w", err)
	}

	cookiePath := filepath.Join(cacheDir, "kusa", "http", "cookies.bin")
	store, err := storage.NewFileStore(cookiePath)
	if err != nil {
		return nil, fmt.Errorf("cannot initialise cookie store: %w", err)
	}

	client := httpclient.New(httpclient.Config{
		BaseURL: baseURL,
		Store:   store,
	})

	sessionPath := filepath.Join(cacheDir, "kusa", "conversation.json")

	return &App{
		Client:  client,
		Session: session.New(sessionPath),
		store:   store,
	}, nil
}

// IsLoggedIn reports whether a session cookie is currently stored.
func (a *App) IsLoggedIn() bool {
	return a.store.IsSet(sessionCookieKey)
}
