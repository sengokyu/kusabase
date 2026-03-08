package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// Storage implements ports.Storage using the local filesystem.
type Storage struct {
	dir string // ストレージのベースディレクトリパス
}

// activeSessionFile は conversation.json の JSON スキーマ。
type activeSessionFile struct {
	ConversationUUID string   `json:"conversationUuid"` // 現在の会話 UUID
	ModelName        string   `json:"modelName"`        // 使用モデル名
	ToolNames        []string `json:"toolNames"`        // 有効化したツール名一覧
}

// sessionsFile は sessions.json の JSON スキーマ。
type sessionsFile struct {
	Sessions []sessionEntry `json:"sessions"` // セッションエントリ一覧
}

// sessionEntry は sessions.json に保存する各セッションのエントリ。
type sessionEntry struct {
	UUID      string   `json:"uuid"`      // 会話の UUID
	ModelName string   `json:"modelName"` // 使用モデル名
	ToolNames []string `json:"toolNames"` // 有効化したツール名一覧
}

// NewStorage creates a Storage backed by dir, creating the directory if needed.
func NewStorage(dir string) (*Storage, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("ストレージディレクトリの作成に失敗しました: %w", err)
	}
	return &Storage{dir: dir}, nil
}

func (s *Storage) activeSessionPath() string {
	return filepath.Join(s.dir, "conversation.json")
}

func (s *Storage) sessionsPath() string {
	return filepath.Join(s.dir, "sessions.json")
}

// SaveActiveSession persists the active session.
func (s *Storage) SaveActiveSession(session domain.ActiveSession) error {
	data := activeSessionFile{
		ConversationUUID: session.ConversationUUID,
		ModelName:        session.ModelName,
		ToolNames:        session.ToolNames,
	}
	return writeJSON(s.activeSessionPath(), data)
}

// LoadActiveSession returns the active session, or nil if none exists.
func (s *Storage) LoadActiveSession() (*domain.ActiveSession, error) {
	var data activeSessionFile
	if err := readJSON(s.activeSessionPath(), &data); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("アクティブセッションの読み込みに失敗しました: %w", err)
	}
	if data.ConversationUUID == "" {
		return nil, nil
	}
	toolNames := data.ToolNames
	if toolNames == nil {
		toolNames = []string{}
	}
	return &domain.ActiveSession{
		ConversationUUID: data.ConversationUUID,
		ModelName:        data.ModelName,
		ToolNames:        toolNames,
	}, nil
}

// ClearActiveSession removes the active session file.
func (s *Storage) ClearActiveSession() error {
	err := os.Remove(s.activeSessionPath())
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("アクティブセッションの削除に失敗しました: %w", err)
	}
	return nil
}

// SaveSessionMeta saves or updates metadata for a session.
func (s *Storage) SaveSessionMeta(meta ports.SessionMeta) error {
	sf, err := s.loadSessions()
	if err != nil {
		sf = &sessionsFile{}
	}

	found := false
	for i, m := range sf.Sessions {
		if m.UUID == meta.UUID {
			sf.Sessions[i] = sessionEntry{
				UUID:      meta.UUID,
				ModelName: meta.ModelName,
				ToolNames: meta.ToolNames,
			}
			found = true
			break
		}
	}
	if !found {
		sf.Sessions = append(sf.Sessions, sessionEntry{
			UUID:      meta.UUID,
			ModelName: meta.ModelName,
			ToolNames: meta.ToolNames,
		})
	}

	return writeJSON(s.sessionsPath(), sf)
}

// LoadSessionsMeta returns all locally stored session metadata.
func (s *Storage) LoadSessionsMeta() ([]ports.SessionMeta, error) {
	sf, err := s.loadSessions()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	metas := make([]ports.SessionMeta, 0, len(sf.Sessions))
	for _, m := range sf.Sessions {
		toolNames := m.ToolNames
		if toolNames == nil {
			toolNames = []string{}
		}
		metas = append(metas, ports.SessionMeta{
			UUID:      m.UUID,
			ModelName: m.ModelName,
			ToolNames: toolNames,
		})
	}
	return metas, nil
}

// DeleteSessionMeta removes the session with the given UUID from local storage.
func (s *Storage) DeleteSessionMeta(uuid string) error {
	sf, err := s.loadSessions()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	filtered := make([]sessionEntry, 0, len(sf.Sessions))
	for _, m := range sf.Sessions {
		if m.UUID != uuid {
			filtered = append(filtered, m)
		}
	}
	sf.Sessions = filtered
	return writeJSON(s.sessionsPath(), sf)
}

func (s *Storage) loadSessions() (*sessionsFile, error) {
	var sf sessionsFile
	if err := readJSON(s.sessionsPath(), &sf); err != nil {
		return nil, err
	}
	return &sf, nil
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON のシリアライズに失敗しました: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗しました: %w", err)
	}
	return nil
}

func readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Ensure Storage satisfies the interface at compile time.
var _ ports.Storage = (*Storage)(nil)
