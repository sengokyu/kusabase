package kusaclient

import "context"

// AuthService provides authentication operations.
type AuthService struct {
	client *Client
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginWithPassword authenticates with the given email and password.
// On success the session cookie is automatically saved via the configured Store.
func (s *AuthService) LoginWithPassword(ctx context.Context, email, password string) error {
	return s.client.doJSON(ctx, "POST", "/api/auth/password/login", loginRequest{
		Email:    email,
		Password: password,
	}, nil)
}

// Probe checks which authentication methods are available for the given email.
func (s *AuthService) Probe(ctx context.Context, req AuthProbeRequest) (*AuthProbeResponse, error) {
	var resp AuthProbeResponse
	if err := s.client.doJSON(ctx, "POST", "/api/auth/probe", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
