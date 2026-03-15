package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/sengokyu/kusabase/cli/internal/app"
	httpclient "github.com/sengokyu/kusabase/httpclient"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the kusa API",
	RunE:  runLogin,
}

func runLogin(_ *cobra.Command, _ []string) error {
	a, err := app.New()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	ctx := context.Background()

	// Probe auth methods; non-fatal if it fails.
	probe, err := a.Client.Auth.Probe(ctx, httpclient.AuthProbeRequest{
		Email:       email,
		RedirectURL: "",
	})
	if err != nil {
		debugLog("auth probe failed: %v", err)
	} else if !probe.AllowPassword {
		return fmt.Errorf("password login is not available for this account")
	}

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	if err := a.Client.Auth.LoginWithPassword(ctx, email, string(passwordBytes)); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println("Login successful.")
	return nil
}
