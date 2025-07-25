/*
Copyright © 2025 Oliver Schrenk
*/
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "infuse",
	Short: "Infuse files into git repositories",
}

func Execute() {
	// hide (but not disable) "completion" feature
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
