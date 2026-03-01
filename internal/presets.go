package core

import "context"

// PresetsService provides preset listing operations.
type PresetsService struct {
	t *Transport
}

func NewPresetsService(t *Transport) *PresetsService { return &PresetsService{t: t} }

// List returns all available presets.
func (s *PresetsService) List(ctx context.Context) ([]Preset, error) {
	var resp PresetListResponse
	if err := s.t.DoJSON(ctx, "GET", "/api/preset", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Presets, nil
}
