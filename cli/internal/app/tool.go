package app

import (
	"context"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// ToolUsecase handles tool-related operations.
type ToolUsecase struct {
	api ports.ExternalAPIClient
}

// NewToolUsecase creates a new ToolUsecase.
func NewToolUsecase(api ports.ExternalAPIClient) *ToolUsecase {
	return &ToolUsecase{api: api}
}

// List returns the available tools from the API.
func (u *ToolUsecase) List(ctx context.Context) ([]domain.Tool, error) {
	return u.api.ListTools(ctx)
}
