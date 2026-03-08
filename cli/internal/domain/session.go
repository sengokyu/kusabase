package domain

// ActiveSession represents the currently active chat session.
type ActiveSession struct {
	ConversationUUID string   // 現在のチャットセッションの UUID
	ModelName        string   // 使用モデル名（ローカル保存）
	ToolNames        []string // 有効化したツール名一覧（ローカル保存）
}
