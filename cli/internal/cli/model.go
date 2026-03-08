package cli

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/sengokyu/kusabase/cli/internal/ports"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "利用可能なモデル一覧を表示する",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newContainer()
		if err != nil {
			return err
		}
		defer c.close()

		models, err := c.model.List(cmd.Context())
		if err != nil {
			if errors.Is(err, ports.ErrNotLoggedIn) {
				printNotLoggedIn()
				return nil
			}
			return err
		}

		fmt.Println("Available Models:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, m := range models {
			suffix := ""
			if m.IsDefault {
				suffix = "  *"
			}
			fmt.Fprintf(w, "- %-20s\t%s%s\n", m.ModelID, m.Name, suffix)
		}
		w.Flush()
		return nil
	},
}
