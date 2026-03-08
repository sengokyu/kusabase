package ports

import "github.com/sengokyu/kusabase/cli/internal/domain"

// SessionMeta stores locally-tracked metadata for a conversation.
type SessionMeta struct {
	UUID      string   // 会話の UUID
	ModelName string   // 使用モデル名
	ToolNames []string // 有効化したツール名一覧
}

// Storage defines the persistence operations for the application.
type Storage interface {
	SaveActiveSession(session domain.ActiveSession) error
	LoadActiveSession() (*domain.ActiveSession, error)
	ClearActiveSession() error
	SaveSessionMeta(meta SessionMeta) error
	LoadSessionsMeta() ([]SessionMeta, error)
	DeleteSessionMeta(uuid string) error
}
