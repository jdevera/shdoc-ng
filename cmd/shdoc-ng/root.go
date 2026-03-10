package main

import (
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
@description, @arg, @option, @exitcode, @example, and more.`,
}
