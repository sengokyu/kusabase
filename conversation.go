package kusaclient

import "context"

// ConversationService provides conversation history operations.
type ConversationService struct {
	client *Client
}

type conversationOverviewResponse struct {
	Success bool `json:"success"`
	Data    struct {
		LatestConversations []Conversation `json:"latestConversations"`
	} `json:"data"`
}

// List returns the latest conversations for the authenticated user.
func (s *ConversationService) List(ctx context.Context) ([]Conversation, error) {
	var resp conversationOverviewResponse
	if err := s.client.doJSON(ctx, "GET", "/api/conversation/overview", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data.LatestConversations, nil
}
