package ports

import (
	"context"

	"github.com/sengokyu/kusabase/cli/internal/domain"
)

// ConfiguredTool holds a tool's UUID and its enabled state for a chat session.
type ConfiguredTool struct {
	UUID    string // ツールの UUID
	Enabled bool   // 有効/無効フラグ
}

// ChatRequest holds parameters for sending a chat message.
type ChatRequest struct {
	Content          string           // ユーザーのメッセージ本文
	ConversationUUID string           // 継続する会話の UUID（新規作成時は空）
	ConfiguredTools  []ConfiguredTool // ツール設定一覧（空の場合は省略）
	FastHeaders      bool             // 会話の最初のメッセージのとき true
	IsRetry          bool             // 同一内容を再送信するとき true
	PresetID         *int             // 使用するプリセット ID（省略時はサーバーのデフォルト）
}

// ExternalAPIClient defines the operations available on the remote API.
type ExternalAPIClient interface {
	Login(ctx context.Context, email, password string) error
	SendChat(ctx context.Context, req ChatRequest) (string, error)
	GetConversations(ctx context.Context) ([]domain.Conversation, error)
	ListTools(ctx context.Context) ([]domain.Tool, error)
	ListPresets(ctx context.Context) ([]domain.Model, error)
}
