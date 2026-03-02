package main

import (
	"bufio"
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
	flag.StringVar(&format, "format", "markdown", "Output format: markdown, json")
	flag.BoolVar(&sortFuncs, "sort", false, "Sort functions alphabetically")
	flag.BoolVar(&showSchema, "schema", false, "Print JSON Schema for --format json output and exit")
	flag.StringVarP(&inputFile, "input", "i", "-", "Input file (- for stdin)")
	flag.StringVarP(&outputFile, "output", "o", "-", "Output file (- for stdout)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `shdoc-ng - Generate Markdown documentation from annotated shell scripts.

Reads a shell script and produces documentation by extracting structured
comment blocks written above shell functions. Supports tags like @description,
@arg, @option, @exitcode, @example, and more.

Usage:
  shdoc-ng [flags]
  shdoc-ng < script.sh > docs.md

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

	scanner := bufio.NewScanner(input)
	parser := NewParser()
	for scanner.Scan() {
		parser.ProcessLine(scanner.Text())
	}

	if sortFuncs {
		sort.Slice(parser.doc.Functions, func(i, j int) bool {
			return parser.doc.Functions[i].Name < parser.doc.Functions[j].Name
		})
	}

	switch format {
	case "markdown", "md":
		fmt.Fprint(output, parser.Render())
	case "json":
		out, err := parser.RenderJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(output, out)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s (supported: markdown, json)\n", format)
		os.Exit(1)
	}
}
