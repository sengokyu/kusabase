package cli

import "github.com/spf13/cobra"

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "ログインする",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newContainer()
		if err != nil {
			return err
		}
		defer c.close()
		return c.login.Run(cmd.Context())
	},
}
