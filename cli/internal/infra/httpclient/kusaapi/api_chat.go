package kusaapi

import (
	"context"
	"fmt"
	"io"

	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// SendChat sends a chat message and returns the plain-text response.
func (c *Client) SendChat(ctx context.Context, req ports.ChatRequest) (string, error) {
	type settings struct {
		Enabled bool `json:"enabled"`
	}
	type configuredTool struct {
		UUID     string   `json:"uuid"`
		Settings settings `json:"settings"`
	}
	type reqBody struct {
		Content          string           `json:"content"`
		ConversationUUID string           `json:"conversationUuid,omitempty"`
		FastHeaders      bool             `json:"fastHeaders,omitempty"`
		IsRetry          bool             `json:"isRetry,omitempty"`
		AttachmentUUIDs  []string         `json:"attachmentUuids,omitempty"`
		ConfiguredTools  []configuredTool `json:"configuredTools,omitempty"`
		PresetID         *int             `json:"presetId,omitempty"`
	}

	body := reqBody{
		Content:          req.Content,
		ConversationUUID: req.ConversationUUID,
		FastHeaders:      req.FastHeaders,
		IsRetry:          req.IsRetry,
		PresetID:         req.PresetID,
	}
	for _, t := range req.ConfiguredTools {
		body.ConfiguredTools = append(body.ConfiguredTools, configuredTool{
			UUID:     t.UUID,
			Settings: settings{Enabled: t.Enabled},
		})
	}

	resp, err := c.post(ctx, "/api/chat", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンスの読み取りに失敗しました: %w", err)
	}
	return string(data), nil
}
