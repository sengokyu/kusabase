package kusaapi

import (
	"context"

	"github.com/sengokyu/kusabase/cli/internal/domain"
)

// ListTools returns the available tools from the API.
func (c *Client) ListTools(ctx context.Context) ([]domain.Tool, error) {
	var result struct {
		Data []struct {
			UUID        string   `json:"uuid"`
			Name        string   `json:"name"`
			DisplayName i18nText `json:"displayName"`
			Description i18nText `json:"description"`
		} `json:"data"`
	}

	if err := c.getJSON(ctx, "/api/tools", &result); err != nil {
		return nil, err
	}

	tools := make([]domain.Tool, 0, len(result.Data))
	for _, t := range result.Data {
		tools = append(tools, domain.Tool{
			UUID:        t.UUID,
			Name:        t.Name,
			DisplayName: t.DisplayName.Resolve(),
			Description: t.Description.Resolve(),
		})
	}
	return tools, nil
}
