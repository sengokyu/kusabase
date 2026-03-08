package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// LoginUsecase handles user authentication.
type LoginUsecase struct {
	api ports.ExternalAPIClient
}

// NewLoginUsecase creates a new LoginUsecase.
func NewLoginUsecase(api ports.ExternalAPIClient) *LoginUsecase {
	return &LoginUsecase{api: api}
}

// Run prompts for credentials and logs in.
func (u *LoginUsecase) Run(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Login ID: ")
	loginId, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("ログインIDの読み取りに失敗しました: %w", err)
	}
	loginId = strings.TrimSpace(loginId)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("パスワードの読み取りに失敗しました: %w", err)
	}
	fmt.Println()
	password := string(passwordBytes)

	if err := u.api.Login(ctx, loginId, password); err != nil {
		return fmt.Errorf("ログインに失敗しました: %w", err)
	}

	fmt.Println("ログインに成功しました。")
	return nil
}
