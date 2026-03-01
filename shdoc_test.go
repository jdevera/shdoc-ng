package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// testCaseMeta holds the metadata read from meta.json.
type testCaseMeta struct {
	Tags []string `json:"tags"`
}

// testCase holds all data for a single conformance test case.
type testCase struct {
	name     string
	dir      string
	input    []byte
	expected []byte
	tags     []string
}

// loadTestCases reads all test cases from testdata/ and returns them grouped by tag.
// Cases without a meta.json or without tags default to "compat".
func loadTestCases(t *testing.T) map[string][]testCase {
	t.Helper()

	casesDir := "testdata"
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("Failed to read cases directory: %v", err)
	}

	grouped := make(map[string][]testCase)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		dir := filepath.Join(casesDir, name)

		inputData, err := os.ReadFile(filepath.Join(dir, "input.sh"))
		if err != nil {
			t.Fatalf("Failed to read input for %s: %v", name, err)
		}

		expectedData, err := os.ReadFile(filepath.Join(dir, "expected.md"))
		if err != nil {
			t.Fatalf("Failed to read expected for %s: %v", name, err)
		}

		tags := readTags(t, dir)

		tc := testCase{
			name:     name,
			dir:      dir,
			input:    inputData,
			expected: expectedData,
			tags:     tags,
		}

		for _, tag := range tags {
			grouped[tag] = append(grouped[tag], tc)
		}
	}

	return grouped
}

// readTags reads meta.json from dir and returns the tags list.
// Returns ["compat"] if meta.json doesn't exist or has no tags.
func readTags(t *testing.T, dir string) []string {
	t.Helper()

	metaPath := filepath.Join(dir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return []string{"compat"}
	}

	var meta testCaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Failed to parse %s: %v", metaPath, err)
	}

	if len(meta.Tags) == 0 {
		return []string{"compat"}
	}

	return meta.Tags
}

// diffOutput produces a line-by-line diff-style error report.
func diffOutput(t *testing.T, testName, actual, expected string) {
	t.Helper()
	t.Errorf("Output mismatch for %s", testName)
	actualLines := strings.Split(actual, "\n")
	expectedLines := strings.Split(expected, "\n")
	maxLines := len(actualLines)
	if len(expectedLines) > maxLines {
		maxLines = len(expectedLines)
	}
	for i := 0; i < maxLines; i++ {
		var aLine, eLine string
		if i < len(actualLines) {
			aLine = actualLines[i]
		}
		if i < len(expectedLines) {
			eLine = expectedLines[i]
		}
		if aLine != eLine {
			t.Errorf("  line %d:\n    expected: %q\n    actual:   %q", i+1, eLine, aLine)
		}
	}
}

func TestConformance(t *testing.T) {
	grouped := loadTestCases(t)

	for tag, cases := range grouped {
		t.Run(tag, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					parser := NewParser()
					lines := strings.Split(string(tc.input), "\n")
					// Remove last element if it's empty (trailing newline in file)
					if len(lines) > 0 && lines[len(lines)-1] == "" {
						lines = lines[:len(lines)-1]
					}
					for _, line := range lines {
						parser.ProcessLine(line)
					}

					actual := parser.Render()
					expected := string(tc.expected)

					if actual != expected {
						diffOutput(t, tc.name, actual, expected)
					}
				})
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	input := `#!/bin/bash
# @name mylib
# @brief A brief description
# @description The full description.

# @section Utils
# @description Helper functions.

# @description Greet someone.
#
# @example
#   greet "World"
#
# @option -u | --uppercase  Uppercase the name.
# @arg $1 string The name to greet.
# @arg $@ string Additional names.
# @set LAST_GREETED string The last name greeted.
# @exitcode 0 Success.
# @exitcode 1 No name provided.
# @stdin A fallback name.
# @stdout The greeting.
# @stderr A warning.
# @see farewell()
# @deprecated Use hello() instead.
greet() {
    echo "Hello, $1!"
}

# @description Say goodbye.
# @arg $1 string The name.
# @noargs
farewell() {
    echo "Bye"
}`

	parser := NewParser()
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		parser.ProcessLine(line)
	}

	jsonOut, err := parser.RenderJSON()
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var doc JSONDocument
	if err := json.Unmarshal([]byte(jsonOut), &doc); err != nil {
		t.Fatalf("Invalid JSON: %v\nOutput:\n%s", err, jsonOut)
	}

	// Check document-level fields
	if doc.Name != "mylib" {
		t.Errorf("Expected name 'mylib', got %q", doc.Name)
	}
	if doc.Brief != "A brief description" {
		t.Errorf("Expected brief 'A brief description', got %q", doc.Brief)
	}
	if doc.Description != "The full description." {
		t.Errorf("Expected description 'The full description.', got %q", doc.Description)
	}

	if len(doc.Functions) != 2 {
		t.Fatalf("Expected 2 functions, got %d", len(doc.Functions))
	}

	// Check first function
	f := doc.Functions[0]
	if f.Name != "greet" {
		t.Errorf("Expected function name 'greet', got %q", f.Name)
	}
	if f.Section != "Utils" {
		t.Errorf("Expected section 'Utils', got %q", f.Section)
	}
	if f.SectionDescription != "Helper functions." {
		t.Errorf("Expected section_description 'Helper functions.', got %q", f.SectionDescription)
	}
	if f.Deprecated != "Use hello() instead." {
		t.Errorf("Expected deprecated 'Use hello() instead.', got %q", f.Deprecated)
	}
	if len(f.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(f.Options))
	}
	if len(f.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(f.Args))
	} else {
		if f.Args[0].Name != "$1" || f.Args[0].Type != "string" {
			t.Errorf("Unexpected arg[0]: %+v", f.Args[0])
		}
		if f.Args[1].Name != "$@" || f.Args[1].Type != "string" {
			t.Errorf("Unexpected arg[1]: %+v", f.Args[1])
		}
	}
	if len(f.Set) != 1 {
		t.Errorf("Expected 1 set var, got %d", len(f.Set))
	} else if f.Set[0].Name != "LAST_GREETED" {
		t.Errorf("Expected set var 'LAST_GREETED', got %q", f.Set[0].Name)
	}
	if len(f.ExitCodes) != 2 {
		t.Errorf("Expected 2 exit codes, got %d", len(f.ExitCodes))
	}
	if len(f.Stdin) != 1 {
		t.Errorf("Expected 1 stdin, got %d", len(f.Stdin))
	}
	if len(f.Stdout) != 1 {
		t.Errorf("Expected 1 stdout, got %d", len(f.Stdout))
	}
	if len(f.Stderr) != 1 {
		t.Errorf("Expected 1 stderr, got %d", len(f.Stderr))
	}
	if len(f.See) != 1 {
		t.Errorf("Expected 1 see, got %d", len(f.See))
	}

	// Check second function
	f2 := doc.Functions[1]
	if f2.Name != "farewell" {
		t.Errorf("Expected function name 'farewell', got %q", f2.Name)
	}
	if !f2.NoArgs {
		t.Errorf("Expected noargs to be true")
	}
}

func TestSortOutput(t *testing.T) {
	input := `#!/bin/bash

# @description Zulu function.
# @noargs
zulu() {
    :
}

# @description Alpha function.
# @noargs
alpha() {
    :
}

# @description Mike function.
# @noargs
mike() {
    :
}`

	parser := NewParser()
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		parser.ProcessLine(line)
	}

	// Sort functions
	sort.Slice(parser.doc.Functions, func(i, j int) bool {
		return parser.doc.Functions[i].Name < parser.doc.Functions[j].Name
	})

	output := parser.Render()

	// Functions should appear in alphabetical order
	alphaIdx := strings.Index(output, "### alpha")
	mikeIdx := strings.Index(output, "### mike")
	zuluIdx := strings.Index(output, "### zulu")

	if alphaIdx == -1 || mikeIdx == -1 || zuluIdx == -1 {
		t.Fatalf("Missing function headers in output:\n%s", output)
	}

	if !(alphaIdx < mikeIdx && mikeIdx < zuluIdx) {
		t.Errorf("Functions not in alphabetical order: alpha@%d, mike@%d, zulu@%d", alphaIdx, mikeIdx, zuluIdx)
	}

	// TOC should also be sorted
	tocAlpha := strings.Index(output, "* [alpha](#alpha)")
	tocMike := strings.Index(output, "* [mike](#mike)")
	tocZulu := strings.Index(output, "* [zulu](#zulu)")

	if tocAlpha == -1 || tocMike == -1 || tocZulu == -1 {
		t.Fatalf("Missing TOC entries in output:\n%s", output)
	}

	if !(tocAlpha < tocMike && tocMike < tocZulu) {
		t.Errorf("TOC not in alphabetical order: alpha@%d, mike@%d, zulu@%d", tocAlpha, tocMike, tocZulu)
	}
}

func TestExternal(t *testing.T) {
	shdocCmd := os.Getenv("SHDOC_CMD")
	if shdocCmd == "" {
		t.Skip("SHDOC_CMD not set")
	}

	parts := strings.Fields(shdocCmd)
	cmdName := parts[0]
	cmdArgs := parts[1:]

	grouped := loadTestCases(t)

	for tag, cases := range grouped {
		t.Run(tag, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					cmd := exec.Command(cmdName, cmdArgs...)
					cmd.Stdin = strings.NewReader(string(tc.input))

					output, err := cmd.Output()
					if err != nil {
						t.Fatalf("Command failed: %v", err)
					}

					actual := string(output)
					expected := string(tc.expected)

					if actual != expected {
						diffOutput(t, tc.name, actual, expected)
					}
				})
			}
		})
	}
}
