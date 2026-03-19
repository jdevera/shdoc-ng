package main

import (
	"fmt"
	"io"
	"os"

	shdoc "github.com/jdevera/shdoc-ng/shdoc"

	"github.com/spf13/cobra"
)

var checkInputFile string

var checkCmd = &cobra.Command{
	Use:   "check [files...]",
	Short: "Check a shell script for documentation warnings",
	Long: `Check shell scripts for documentation warnings without generating output.
Exits with code 1 if any warnings are found.

Files can be passed as positional arguments, via -i, or on stdin.

Examples:
  shdoc-ng check script.sh lib.sh
  shdoc-ng check -i script.sh
  shdoc-ng check < script.sh`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&checkInputFile, "input", "i", "", "Input file (deprecated, use positional args)")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Build file list: -i flag, positional args, or stdin
	files := args
	if checkInputFile != "" {
		files = append([]string{checkInputFile}, files...)
	}

	if len(files) == 0 {
		return checkFile("-")
	}

	var totalWarns int
	for _, f := range files {
		if err := checkFile(f); err != nil {
			totalWarns++
		}
	}

	if totalWarns > 0 {
		return fmt.Errorf("%d file(s) with warnings", totalWarns)
	}
	return nil
}

func checkFile(path string) error {
	var input io.Reader
	warnFile := path
	if path == "-" {
		input = os.Stdin
		warnFile = "<stdin>"
	} else {
		f, err := os.Open(path)
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

	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "%s:%d:%d: warning: %s\n", warnFile, w.Line, w.Col+1, w.Message)
	}

	if len(warns) > 0 {
		return fmt.Errorf("found %d warning(s)", len(warns))
	}

	return nil
}
