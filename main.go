package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	flag "github.com/spf13/pflag"
)

func main() {
	var format string
	var sortFuncs bool
	var showSchema bool
	var inputFile string
	var outputFile string
	var printTemplate string
	var templateFile string
	flag.StringVar(&format, "format", "markdown", "Output format: markdown, html, json")
	flag.BoolVar(&sortFuncs, "sort", false, "Sort functions alphabetically")
	flag.BoolVar(&showSchema, "schema", false, "Print JSON Schema for --format json output and exit")
	flag.StringVarP(&inputFile, "input", "i", "-", "Input file (- for stdin)")
	flag.StringVarP(&outputFile, "output", "o", "-", "Output file (- for stdout)")
	flag.StringVar(&printTemplate, "print-template", "", "Print built-in template for format (markdown, html) and exit")
	flag.StringVar(&templateFile, "template", "", "Use a custom template file instead of the built-in one for the selected --format")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `shdoc-ng - Generate documentation from annotated shell scripts.

Reads a shell script and produces documentation by extracting structured
comment blocks written above shell functions. Supports tags like @description,
@arg, @option, @exitcode, @example, and more.

Usage:
  shdoc-ng [flags]
  shdoc-ng < script.sh > docs.md
  shdoc-ng --format html -i script.sh -o docs.html

Flags:
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	var output io.Writer
	if outputFile == "-" {
		output = os.Stdout
	} else {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		output = f
	}

	if showSchema {
		out, err := renderSchema()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating schema: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(output, out)
		return
	}

	if printTemplate != "" {
		switch printTemplate {
		case "md", "markdown":
			fmt.Fprint(output, defaultMarkdownTemplate)
		case "html":
			fmt.Fprint(output, defaultHTMLTemplate)
		default:
			fmt.Fprintf(os.Stderr, "Unknown template format: %q (supported: markdown, html)\n", printTemplate)
			os.Exit(1)
		}
		return
	}

	var input io.Reader
	if inputFile == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
	}

	src, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	doc, warns := ParseDocument(string(src))

	warnColor := "\033[1;34m"
	colorClear := "\033[1;0m"
	for _, w := range warns {
		fmt.Fprintf(os.Stderr, "%sline %4d, warning : %s%s\n", warnColor, w.Line, w.Message, colorClear)
	}

	if sortFuncs {
		sort.Slice(doc.Functions, func(i, j int) bool {
			return doc.Functions[i].Name < doc.Functions[j].Name
		})
	}

	switch format {
	case "markdown", "md":
		tmplText := defaultMarkdownTemplate
		if templateFile != "" {
			data, err := os.ReadFile(templateFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading template file: %v\n", err)
				os.Exit(1)
			}
			tmplText = string(data)
		}
		out, err := renderWithTemplate(&doc, tmplText)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering markdown: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(output, out)
	case "html":
		tmplText := defaultHTMLTemplate
		if templateFile != "" {
			data, err := os.ReadFile(templateFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading template file: %v\n", err)
				os.Exit(1)
			}
			tmplText = string(data)
		}
		out, err := renderWithTemplate(&doc, tmplText)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering HTML: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(output, out)
	case "json":
		out, err := renderDocumentJSON(&doc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(output, out)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %q (supported: markdown, html, json)\n", format)
		os.Exit(1)
	}
}
