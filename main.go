package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
)

func main() {
	var format string
	var sortFuncs bool
	var showSchema bool
	flag.StringVar(&format, "format", "markdown", "Output format: markdown, json")
	flag.BoolVar(&sortFuncs, "sort", false, "Sort functions alphabetically")
	flag.BoolVar(&showSchema, "schema", false, "Print JSON Schema for --format json output and exit")
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

	scanner := bufio.NewScanner(os.Stdin)
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
		fmt.Print(parser.Render())
	case "json":
		out, err := parser.RenderJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(out)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s (supported: markdown, json)\n", format)
		os.Exit(1)
	}
}
