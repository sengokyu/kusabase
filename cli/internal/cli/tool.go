package cli

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/sengokyu/kusabase/cli/internal/ports"
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "利用可能なツール一覧を表示する",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newContainer()
		if err != nil {
			return err
		}
		defer c.close()

		tools, err := c.tool.List(cmd.Context())
		if err != nil {
			if errors.Is(err, ports.ErrNotLoggedIn) {
				printNotLoggedIn()
				return nil
			}
			return err
		}

		fmt.Println("Available Tools:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, t := range tools {
			desc := t.Description
			if desc == "" {
				desc = t.DisplayName
			}
			fmt.Fprintf(w, "- %-20s\t: %s\n", t.Name, desc)
		}
		w.Flush()
		return nil
	},
}
