package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session persists the active conversation UUID between CLI invocations.
// Stored as JSON at XDG Cache Home/kusa/conversation.json.
type Session struct {
	path string
}

type sessionData struct {
	ConversationUUID string `json:"conversationUUID"`
}

// New returns a Session backed by the file at path.
func New(path string) *Session {
	return &Session{path: path}
}

// Load returns the stored conversation UUID, or ("", err) on failure.
func (s *Session) Load() (string, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return "", err
	}
	var d sessionData
	if err := json.Unmarshal(b, &d); err != nil {
		return "", err
	}
	return d.ConversationUUID, nil
}

// Save persists the conversation UUID to disk.
func (s *Session) Save(conversationUUID string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	b, err := json.Marshal(sessionData{ConversationUUID: conversationUUID})
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0600)
}
