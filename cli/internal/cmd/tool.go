package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sengokyu/kusabase/cli/internal/app"
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "List available tools",
	RunE:  runTool,
}

func runTool(_ *cobra.Command, _ []string) error {
	a, err := app.New()
	if err != nil {
		return err
	}

	if !a.IsLoggedIn() {
		printNotLoggedIn()
		return nil
	}

	ctx := context.Background()
	tools, err := a.Client.Tools.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Println("Available Tools:")
	for _, t := range tools {
		desc := t.Description["en"]
		if desc == "" {
			desc = t.Description["ja"]
		}
		fmt.Printf("- %-20s : %s\n", t.Name, desc)
	}
	return nil
}
