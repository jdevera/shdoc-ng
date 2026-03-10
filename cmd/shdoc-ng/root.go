package main

import (
	"os"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "shdoc-ng",
	Version: version,
	Short:   "Generate documentation from annotated shell scripts",
	Long: `shdoc-ng reads shell scripts and produces documentation by extracting
structured comment blocks written above shell functions. Supports tags like
@description, @arg, @option, @exitcode, @example, and more.

When invoked without a subcommand and stdin is a pipe, it behaves as
"shdoc-ng generate" for convenient use in pipelines:

  shdoc-ng < script.sh > docs.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If stdin is a pipe, behave as "generate".
		fi, _ := os.Stdin.Stat()
		if fi.Mode()&os.ModeCharDevice == 0 {
			return generateCmd.RunE(generateCmd, args)
		}
		return cmd.Help()
	},
}
