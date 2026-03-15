package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "kusa",
	Short: "kusa – AI chat CLI",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "output debug logs to STDERR")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(modelCmd)
	rootCmd.AddCommand(toolCmd)
}

func debugLog(format string, args ...any) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

func printNotLoggedIn() {
	fmt.Println("You are not logged in.")
	fmt.Println("Run `kusa login` to continue.")
}
