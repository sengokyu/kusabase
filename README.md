# Unofficial HTTP client library for Kusa

Unofficial Go HTTP client for the Kusa GAI API.

This library wraps the REST endpoints defined in the OpenAPI specification and
provides:

- Cookie-based session management
- Typed request/response models
- Simple abstractions for Auth / Conversation / Tools / Chat / Presets

> Note: This is an unofficial client. The API may change without notice.

---

## Installation

```bash
go get github.com/sengokyu/kusaclient
```

---

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/sengokyu/kusaclient"
)

// Implement a persistent storage for session cookies, tokens, etc.
struct Store { /* */ }
func (s *Store) Save(ctx context, key string, value string) error { /* */ }
func (s *Store) Load(ctx context, key string) (string, error) { /* */ }

// Ensure Store implements kusaclient.Store
var _ kusaclient.Store = (*Store)(nil)

func main() {
    client := kusaclient.New(kusaclient.Config{
        BaseURL: "https://gai.example.com",
        Store:   Store,
    })

    ctx := context.Background()

    // Authenticate
    err := client.Auth.LoginWithPassword(ctx, "user@example.com", "password")

    // Get model list
    presets, err := client.Presets.List(ctx)

    // Start new conversation
    chat, err := client.Chat.New(presets[0], kusaclient.ChatRequest{
        Content: "Hello"
    })
    chat.Send(kusaclient.ChatRequest{
        Content: "How are you"
    })
}
```

## Error Handling

All non-2xx responses are converted into `APIError`:

```go
type APIError struct {
    StatusCode int
    Namespace  string
    Key        string
}
```

Example:

```go
if err != nil {
    if apiErr, ok := err.(*kusaclient.APIError); ok {
        log.Printf("api error: %d %s:%s", apiErr.StatusCode, apiErr.Namespace, apiErr.Key)
    }
}
```

---

## Thread Safety

`Client` is safe for concurrent use.  
Your `Store` implementation must be thread-safe if used concurrently.

---

## License

MIT
