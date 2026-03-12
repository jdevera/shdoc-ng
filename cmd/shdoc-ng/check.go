package main

import (
	"fmt"
	"io"
	"os"

	shdoc "github.com/jdevera/shdoc-ng"

	"github.com/spf13/cobra"
)

var checkInputFile string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check a shell script for documentation warnings",
	Long: `Check a shell script for documentation warnings without generating output.
Exits with code 1 if any warnings are found.

Examples:
  shdoc-ng check -i script.sh
  shdoc-ng check < script.sh`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&checkInputFile, "input", "i", "-", "Input file (- for stdin)")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	var input io.Reader
	if checkInputFile == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(checkInputFile)
		if err != nil {
			return fmt.Errorf("opening input file: %w", err)
		}
		defer func() { _ = f.Close() }()
		input = f
	}

	src, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	_, warns := shdoc.ParseDocument(string(src))

	warnFile := checkInputFile
	if warnFile == "-" {
		warnFile = "<stdin>"
	}
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "%s:%d:%d: warning: %s\n", warnFile, w.Line, w.Col+1, w.Message)
	}

	if len(warns) > 0 {
		return fmt.Errorf("found %d warning(s)", len(warns))
	}

	return nil
}
