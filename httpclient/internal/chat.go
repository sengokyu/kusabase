package core

import "context"

// ChatService provides chat operations.
type ChatService struct {
	t *Transport
}

func NewChatService(t *Transport) *ChatService { return &ChatService{t: t} }

// Chat represents an ongoing chat conversation with a specific preset.
type Chat struct {
	t        *Transport
	presetID int

	// LastResponse holds the AI reply from the most recent Send (or New) call.
	LastResponse string
}

type chatAPIRequest struct {
	AttachmentUUIDs  []string         `json:"attachmentUuids,omitempty"`
	ConfiguredTools  []ConfiguredTool `json:"configuredTools,omitempty"`
	Content          string           `json:"content"`
	ConversationUUID string           `json:"conversationUuid,omitempty"`
	FastHeaders      bool             `json:"fastHeaders,omitempty"`
	IsRetry          bool             `json:"isRetry,omitempty"`
	PresetID         int              `json:"presetId,omitempty"`
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
	reply, err := s.t.PostText(context.Background(), "/api/chat", apiReq)
	if err != nil {
		return nil, err
	}
	return &Chat{
		t:            s.t,
		presetID:     preset.ID,
		LastResponse: reply,
	}, nil
}

// Send sends a chat message in a stateless manner (no conversation state is tracked).
// ConversationUUID, FastHeaders and PresetID in req are forwarded as-is.
func (s *ChatService) Send(ctx context.Context, req ChatRequest) (string, error) {
	var presetID int
	if req.PresetID != nil {
		presetID = *req.PresetID
	}
	apiReq := chatAPIRequest{
		Content:          req.Content,
		AttachmentUUIDs:  req.AttachmentUUIDs,
		ConfiguredTools:  req.ConfiguredTools,
		IsRetry:          req.IsRetry,
		FastHeaders:      req.FastHeaders,
		ConversationUUID: req.ConversationUUID,
		PresetID:         presetID,
	}
	return s.t.PostText(ctx, "/api/chat", apiReq)
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
	reply, err := c.t.PostText(context.Background(), "/api/chat", apiReq)
	if err != nil {
		return "", err
	}
	c.LastResponse = reply
	return reply, nil
}
