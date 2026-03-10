package main

import (
	"github.com/jdevera/shdoc-ng/internal/lsp"

	"github.com/spf13/cobra"
)

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Run the LSP server on stdio",
	Long: `Run the shdoc-ng Language Server Protocol server. Communicates over stdio
and provides diagnostics, hover, completion, go-to-definition, document symbols,
folding ranges, and code actions for shell script documentation.

Editor configuration:
  VSCode: set "shdoc-ng.serverCommand" to "shdoc-ng lsp"
  Neovim: require('shdoc-ng').setup({ cmd = 'shdoc-ng', args = { 'lsp' } })`,
	Run: func(cmd *cobra.Command, args []string) {
		lsp.Run()
	},
}

func init() {
	rootCmd.AddCommand(lspCmd)
}
