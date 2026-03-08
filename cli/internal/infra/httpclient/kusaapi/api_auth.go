package kusaapi

import (
	"context"
)

// Login authenticates with email and password.
// 認証失敗時はサーバーが HTTP 400 を返す（仕様 /api/auth/password/login 参照）。
func (c *Client) Login(ctx context.Context, email, password string) error {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	resp, err := c.post(ctx, "/api/auth/password/login", body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
