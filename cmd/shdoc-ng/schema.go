package main

import (
	"fmt"

	shdoc "github.com/jdevera/shdoc-ng"

	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "json-schema",
	Short: "Print the JSON Schema for the JSON output format",
	RunE:  runSchema,
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}

func runSchema(cmd *cobra.Command, args []string) error {
	out, err := shdoc.RenderSchema()
	if err != nil {
		return fmt.Errorf("generating schema: %w", err)
	}
	fmt.Print(out)
	return nil
}
