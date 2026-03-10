package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	shdoc "github.com/jdevera/shdoc-ng"

	"github.com/spf13/cobra"
)

var (
	genFormat       string
	genSort         bool
	genInputFile    string
	genOutputFile   string
	genTemplateFile string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation from a shell script",
	Long: `Generate documentation from annotated shell scripts. Reads a shell script
and produces Markdown, HTML, or JSON output.

Examples:
  shdoc-ng generate -i script.sh -o docs.md
  shdoc-ng generate --format html -i script.sh -o docs.html
  shdoc-ng generate < script.sh > docs.md`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&genFormat, "format", "markdown", "Output format: markdown, html, json")
	generateCmd.Flags().BoolVar(&genSort, "sort", false, "Sort functions alphabetically")
	generateCmd.Flags().StringVarP(&genInputFile, "input", "i", "-", "Input file (- for stdin)")
	generateCmd.Flags().StringVarP(&genOutputFile, "output", "o", "-", "Output file (- for stdout)")
	generateCmd.Flags().StringVar(&genTemplateFile, "template", "", "Use a custom template file instead of the built-in one")
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	var output io.Writer
	if genOutputFile == "-" {
		output = os.Stdout
	} else {
		f, err := os.Create(genOutputFile)
		if err != nil {
			return fmt.Errorf("opening output file: %w", err)
		}
		defer f.Close()
		output = f
	}

	var input io.Reader
	if genInputFile == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(genInputFile)
		if err != nil {
			return fmt.Errorf("opening input file: %w", err)
		}
		defer f.Close()
		input = f
	}

	src, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	doc, warns := shdoc.ParseDocument(string(src))

	warnFile := genInputFile
	if warnFile == "-" {
		warnFile = "<stdin>"
	}
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "%s:%d:%d: warning: %s\n", warnFile, w.Line, w.Col+1, w.Message)
	}

	if genSort {
		sort.SliceStable(doc.Functions, func(i, j int) bool {
			if doc.Functions[i].Section != doc.Functions[j].Section {
				return false // preserve section order
			}
			return doc.Functions[i].Name < doc.Functions[j].Name
		})
		// Recompute IsFirstInSection after reordering.
		seenSections := map[string]bool{}
		for i := range doc.Functions {
			f := &doc.Functions[i]
			f.IsFirstInSection = f.Section != "" && !seenSections[f.Section]
			if f.Section != "" {
				seenSections[f.Section] = true
			}
		}
	}

	switch genFormat {
	case "markdown", "md":
		tmplText := shdoc.DefaultMarkdownTemplate
		if genTemplateFile != "" {
			data, err := os.ReadFile(genTemplateFile)
			if err != nil {
				return fmt.Errorf("reading template file: %w", err)
			}
			tmplText = string(data)
		}
		out, err := shdoc.RenderWithTemplate(&doc, tmplText)
		if err != nil {
			return fmt.Errorf("rendering markdown: %w", err)
		}
		fmt.Fprint(output, out)
	case "html":
		tmplText := shdoc.DefaultHTMLTemplate
		if genTemplateFile != "" {
			data, err := os.ReadFile(genTemplateFile)
			if err != nil {
				return fmt.Errorf("reading template file: %w", err)
			}
			tmplText = string(data)
		}
		out, err := shdoc.RenderWithTemplate(&doc, tmplText)
		if err != nil {
			return fmt.Errorf("rendering HTML: %w", err)
		}
		fmt.Fprint(output, out)
	case "json":
		out, err := shdoc.RenderDocumentJSON(&doc)
		if err != nil {
			return fmt.Errorf("rendering JSON: %w", err)
		}
		fmt.Fprint(output, out)
	default:
		return fmt.Errorf("unknown format: %q (supported: markdown, html, json)", genFormat)
	}

	return nil
}
