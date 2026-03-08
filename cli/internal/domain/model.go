package domain

// Model represents an available AI model backed by a server preset.
type Model struct {
	ID        int    // Preset ID（chat API の presetId に使用）
	UUID      string // Preset UUID
	Name      string // 表示名（日本語優先）
	ModelID   string // モデル識別子（例: "gpt-4.1"）
	IsDefault bool   // デフォルトプリセットのとき true
}
