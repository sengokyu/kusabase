package kusaclient

import "context"

// ToolsService provides tool listing operations.
type ToolsService struct {
	client *Client
}

type toolsResponse struct {
	Success bool   `json:"success"`
	Data    []Tool `json:"data"`
}

// List returns all available tools.
func (s *ToolsService) List(ctx context.Context) ([]Tool, error) {
	var resp toolsResponse
	if err := s.client.doJSON(ctx, "GET", "/api/tools", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
