package kusaclient

import "context"

// PresetsService provides preset listing operations.
type PresetsService struct {
	client *Client
}

// List returns all available presets.
func (s *PresetsService) List(ctx context.Context) ([]Preset, error) {
	var resp PresetListResponse
	if err := s.client.doJSON(ctx, "GET", "/api/preset", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Presets, nil
}
