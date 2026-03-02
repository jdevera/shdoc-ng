package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
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
	flag.StringVar(&inputFile, "i", "-", "Input file (- for stdin)")
	flag.StringVar(&outputFile, "o", "-", "Output file (- for stdout)")
	flag.Parse()

	if showSchema {
		out, err := renderSchema()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating schema: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(out)
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
