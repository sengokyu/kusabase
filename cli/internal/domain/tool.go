package domain

// Tool represents an available AI tool.
type Tool struct {
	UUID        string // ツールの UUID
	Name        string // ツール識別名（例: web_search）
	DisplayName string // 表示名（日本語優先）
	Description string // ツールの説明（日本語優先）
}
