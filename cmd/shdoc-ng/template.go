package main

import (
	"fmt"

	shdoc "github.com/jdevera/shdoc-ng"

	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:       "template [markdown|html]",
	Short:     "Print a built-in template",
	Long:      "Print the built-in template for the given format. Useful as a starting point for custom templates.",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"markdown", "md", "html"},
	RunE:      runTemplate,
}

func init() {
	rootCmd.AddCommand(templateCmd)
}

func runTemplate(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "md", "markdown":
		fmt.Print(shdoc.DefaultMarkdownTemplate)
	case "html":
		fmt.Print(shdoc.DefaultHTMLTemplate)
	default:
		return fmt.Errorf("unknown template format: %q (supported: markdown, html)", args[0])
	}
	return nil
}
