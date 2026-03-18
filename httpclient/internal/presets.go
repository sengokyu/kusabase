package core

import "context"

// PresetsService provides preset listing operations.
type PresetsService struct {
	transport *Transport
}

func NewPresetsService(t *Transport) *PresetsService {
	return &PresetsService{transport: t}
}

// List returns all available presets and the default preset UUID.
func (s *PresetsService) List(ctx context.Context) (PresetListResponse, error) {
	var resp PresetListResponse
	if err := s.transport.DoJSON(ctx, "GET", "/api/preset", nil, &resp); err != nil {
		return PresetListResponse{}, err
	}
	return resp, nil
}
