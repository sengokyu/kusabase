package kusaclient

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// Client is the main entry point for the kusaclient library.
// It is safe for concurrent use provided the Store implementation is also safe.
type Client struct {
	Auth         *AuthService
	Conversation *ConversationService
	Tools        *ToolsService
	Chat         *ChatService
	Presets      *PresetsService

	httpClient *http.Client
	baseURL    string
	store      Store
}

// Config holds the configuration for creating a Client.
type Config struct {
	// BaseURL is the root URL of the Kusa GAI API (e.g. "https://gai.example.com").
	BaseURL string
	// Store is used to persist and restore the session cookie.
	Store Store
}

// New creates a new Client from the given Config.
// If the Store contains a previously saved session, it is restored automatically.
func New(cfg Config) *Client {
	jar, _ := cookiejar.New(nil)

	c := &Client{
		httpClient: &http.Client{Jar: jar},
		baseURL:    cfg.BaseURL,
		store:      cfg.Store,
	}

	// Restore persisted session cookie into the jar.
	ctx := context.Background()
	if session, err := cfg.Store.Load(ctx, "next-session"); err == nil && session != "" {
		if u, err := url.Parse(cfg.BaseURL); err == nil {
			jar.SetCookies(u, []*http.Cookie{{
				Name:  "next-session",
				Value: session,
			}})
		}
	}

	c.Auth = &AuthService{client: c}
	c.Conversation = &ConversationService{client: c}
	c.Tools = &ToolsService{client: c}
	c.Chat = &ChatService{client: c}
	c.Presets = &PresetsService{client: c}

	return c
}
