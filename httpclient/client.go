package httpclient

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	core "github.com/sengokyu/kusabase/httpclient/internal"
)

// Client is the main entry point for the kusaclient library.
// It is safe for concurrent use provided the Store implementation is also safe.
type Client struct {
	Auth         *core.AuthService
	Conversation *core.ConversationService
	Tools        *core.ToolsService
	Chat         *core.ChatService
	Presets      *core.PresetsService
}

// Config holds the configuration for creating a Client.
type Config struct {
	// BaseURL is the root URL of the Kusa GAI API (e.g. "https://gai.example.com").
	BaseURL string
	// Store is used to persist and restore the session cookie.
	CookieStore core.CookieStore
}

// New creates a new Client from the given Config.
// If the Store contains a previously saved session, it is restored automatically.
func New(cfg Config) *Client {
	jar, _ := cookiejar.New(nil)
	hc := &http.Client{Jar: jar}

	// Restore persisted cookies into the jar.
	ctx := context.Background()
	if cookies, err := cfg.CookieStore.Load(ctx); err == nil && len(cookies) > 0 {
		if u, err := url.Parse(cfg.BaseURL); err == nil {
			jar.SetCookies(u, cookies)
		}
	}

	t := core.NewTransport(hc, cfg.BaseURL, func(ctx context.Context, cookies []*http.Cookie) {
		_ = cfg.CookieStore.Save(ctx, cookies)
	})

	return &Client{
		Auth:         core.NewAuthService(t),
		Conversation: core.NewConversationService(t),
		Tools:        core.NewToolsService(t),
		Chat:         core.NewChatService(t),
		Presets:      core.NewPresetsService(t),
	}
}
