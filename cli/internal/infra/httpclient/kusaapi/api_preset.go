package kusaapi

import (
	"context"

	"github.com/sengokyu/kusabase/cli/internal/domain"
)

// ListPresets returns the available presets (models) from the API.
func (c *Client) ListPresets(ctx context.Context) ([]domain.Model, error) {
	var result struct {
		Presets []struct {
			ID              int      `json:"id"`
			UUID            string   `json:"uuid"`
			Name            i18nText `json:"name"`
			ModelParameters struct {
				Model string `json:"model"`
			} `json:"modelParameters"`
		} `json:"presets"`
		DefaultPreset string `json:"defaultPreset"`
	}

	if err := c.getJSON(ctx, "/api/preset", &result); err != nil {
		return nil, err
	}

	models := make([]domain.Model, 0, len(result.Presets))
	for _, p := range result.Presets {
		models = append(models, domain.Model{
			ID:        p.ID,
			UUID:      p.UUID,
			Name:      p.Name.Resolve(),
			ModelID:   p.ModelParameters.Model,
			IsDefault: p.UUID == result.DefaultPreset,
		})
	}
	return models, nil
}
