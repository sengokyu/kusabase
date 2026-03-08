package app

import (
	"context"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// ModelUsecase handles model-related operations.
type ModelUsecase struct {
	api ports.ExternalAPIClient
}

// NewModelUsecase creates a new ModelUsecase.
func NewModelUsecase(api ports.ExternalAPIClient) *ModelUsecase {
	return &ModelUsecase{api: api}
}

// List returns the available models from the API.
func (u *ModelUsecase) List(ctx context.Context) ([]domain.Model, error) {
	return u.api.ListPresets(ctx)
}
