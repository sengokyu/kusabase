package kusaapi

import (
	"context"
	"time"

	"github.com/sengokyu/kusabase/cli/internal/domain"
)

// GetConversations returns the latest conversations from the overview API.
func (c *Client) GetConversations(ctx context.Context) ([]domain.Conversation, error) {
	var result struct {
		Data struct {
			LatestConversations []struct {
				UUID          string `json:"uuid"`
				Title         string `json:"title"`
				LastMessageAt string `json:"lastMessageAt"`
			} `json:"latestConversations"`
		} `json:"data"`
	}

	if err := c.getJSON(ctx, "/api/conversation/overview?page_size=50", &result); err != nil {
		return nil, err
	}

	convs := make([]domain.Conversation, 0, len(result.Data.LatestConversations))
	for _, lc := range result.Data.LatestConversations {
		t, _ := time.Parse(time.RFC3339, lc.LastMessageAt)
		convs = append(convs, domain.Conversation{
			UUID:          lc.UUID,
			Title:         lc.Title,
			LastMessageAt: t,
		})
	}
	return convs, nil
}
