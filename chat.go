package kusaclient

import "context"

// ChatService provides chat operations.
type ChatService struct {
	client *Client
}

// Chat represents an ongoing chat conversation with a specific preset.
type Chat struct {
	client   *Client
	presetID int

	// LastResponse holds the AI reply from the most recent Send (or New) call.
	LastResponse string
}

// chatAPIRequest is the internal request body for POST /api/chat.
type chatAPIRequest struct {
	AttachmentUUIDs []string         `json:"attachmentUuids,omitempty"`
	ConfiguredTools []ConfiguredTool `json:"configuredTools,omitempty"`
	Content         string           `json:"content"`
	FastHeaders     bool             `json:"fastHeaders,omitempty"`
	IsRetry         bool             `json:"isRetry,omitempty"`
	PresetID        int              `json:"presetId,omitempty"`
}

// New starts a new conversation by sending the first message with the given preset.
// The AI reply is stored in Chat.LastResponse.
func (s *ChatService) New(preset Preset, req ChatRequest) (*Chat, error) {
	apiReq := chatAPIRequest{
		Content:         req.Content,
		AttachmentUUIDs: req.AttachmentUUIDs,
		ConfiguredTools: req.ConfiguredTools,
		IsRetry:         req.IsRetry,
		FastHeaders:     true,
		PresetID:        preset.ID,
	}
	reply, err := s.client.doText(context.Background(), "POST", "/api/chat", apiReq)
	if err != nil {
		return nil, err
	}
	return &Chat{
		client:       s.client,
		presetID:     preset.ID,
		LastResponse: reply,
	}, nil
}

// Send sends a follow-up message in the conversation.
// The AI reply is stored in Chat.LastResponse and also returned directly.
func (c *Chat) Send(req ChatRequest) (string, error) {
	apiReq := chatAPIRequest{
		Content:         req.Content,
		AttachmentUUIDs: req.AttachmentUUIDs,
		ConfiguredTools: req.ConfiguredTools,
		IsRetry:         req.IsRetry,
		PresetID:        c.presetID,
	}
	reply, err := c.client.doText(context.Background(), "POST", "/api/chat", apiReq)
	if err != nil {
		return "", err
	}
	c.LastResponse = reply
	return reply, nil
}
