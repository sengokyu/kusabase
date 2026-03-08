package domain

import "time"

// Conversation represents a chat conversation.
type Conversation struct {
	UUID          string    // 会話の UUID
	Title         string    // 会話タイトル
	LastMessageAt time.Time // 最終メッセージ日時
	// ローカル保存のメタデータ（API レスポンスには含まれない）
	ModelName string   // 使用モデル名
	ToolNames []string // 有効化したツール名一覧
}
