package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sengokyu/kusabase/cli/internal/app"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "List available models",
	RunE:  runModel,
}

func runModel(_ *cobra.Command, _ []string) error {
	a, err := app.New()
	if err != nil {
		return err
	}

	if !a.IsLoggedIn() {
		printNotLoggedIn()
		return nil
	}

	ctx := context.Background()
	resp, err := a.Client.Presets.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	seen := make(map[string]bool)
	fmt.Println("Available Models:")
	for _, p := range resp.Presets {
		m := p.ModelParameters.Model
		if m != "" && !seen[m] {
			seen[m] = true
			fmt.Printf("- %s\n", m)
		}
	}
	return nil
}
