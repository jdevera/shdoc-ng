package shdoc

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testCaseMeta holds the metadata read from meta.json.
type testCaseMeta struct {
	Tags             []string `json:"tags"`
	KnownDeviations  []string `json:"knownDeviations,omitempty"`
}

// testCase holds all data for a single conformance test case.
type testCase struct {
	name             string
	dir              string
	input            []byte
	expected         []byte
	tags             []string
	knownDeviations  []string
}

// loadTestCases reads all test cases from testdata/ and returns them grouped by tag.
// Cases without a meta.json or without tags default to "compat".
// When legacy is false, expected-ng.md is preferred over expected.md.
// When legacy is true, only expected.md is used.
func loadTestCases(t *testing.T, legacy bool) map[string][]testCase {
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

		var expectedData []byte
		if !legacy {
			// Prefer expected-ng.md (our corrected output) over expected.md.
			expectedData, err = os.ReadFile(filepath.Join(dir, "expected-ng.md"))
			if err != nil {
				expectedData = nil
			}
		}
		if expectedData == nil {
			expectedData, err = os.ReadFile(filepath.Join(dir, "expected.md"))
			if err != nil {
				t.Fatalf("Failed to read expected for %s: %v", name, err)
			}
		}

		meta := readMeta(t, dir)

		tc := testCase{
			name:            name,
			dir:             dir,
			input:           inputData,
			expected:        expectedData,
			tags:            meta.Tags,
			knownDeviations: meta.KnownDeviations,
		}

		for _, tag := range meta.Tags {
			grouped[tag] = append(grouped[tag], tc)
		}
	}

	return grouped
}

// readMeta reads meta.json from dir and returns the parsed metadata.
// Returns default metadata with tags=["compat"] if meta.json doesn't exist.
func readMeta(t *testing.T, dir string) testCaseMeta {
	t.Helper()

	metaPath := filepath.Join(dir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return testCaseMeta{Tags: []string{"compat"}}
	}

	var meta testCaseMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Failed to parse %s: %v", metaPath, err)
	}

	if len(meta.Tags) == 0 {
		meta.Tags = []string{"compat"}
	}

	return meta
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
	grouped := loadTestCases(t, false)

	for tag, cases := range grouped {
		t.Run(tag, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					if len(tc.knownDeviations) > 0 {
						t.Logf("Known deviations from original shdoc:")
						for _, d := range tc.knownDeviations {
							t.Logf("  - %s", d)
						}
					}
					doc, _ := ParseDocument(string(tc.input))
					actual, err := RenderWithTemplate(&doc, DefaultMarkdownTemplate)
					if err != nil {
						t.Fatalf("RenderWithTemplate() error: %v", err)
					}
					expected := string(tc.expected)
					if actual != expected {
						diffOutput(t, tc.name, actual, expected)
					}
				})
			}
		})
	}
}

// TestLegacyConformance runs the conformance suite using only expected.md
// (the original awk output). Cases with knownDeviations are skipped because
// shdoc-ng intentionally deviates from the original awk behavior.
func TestLegacyConformance(t *testing.T) {
	grouped := loadTestCases(t, true)

	for tag, cases := range grouped {
		t.Run(tag, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					if len(tc.knownDeviations) > 0 {
						t.Skipf("Skipped: known deviations from original shdoc: %v", tc.knownDeviations)
					}
					doc, _ := ParseDocument(string(tc.input))
					actual, err := RenderWithTemplate(&doc, DefaultMarkdownTemplate)
					if err != nil {
						t.Fatalf("RenderWithTemplate() error: %v", err)
					}
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

	doc, _ := ParseDocument(input)

	jsonOut, err := RenderDocumentJSON(&doc)
	if err != nil {
		t.Fatalf("renderDocumentJSON failed: %v", err)
	}

	var parsedDoc Document
	if err := json.Unmarshal([]byte(jsonOut), &parsedDoc); err != nil {
		t.Fatalf("Invalid JSON: %v\nOutput:\n%s", err, jsonOut)
	}

	// Check document-level fields
	if parsedDoc.FileTitle != "mylib" {
		t.Errorf("Expected name 'mylib', got %q", parsedDoc.FileTitle)
	}
	if parsedDoc.FileBrief != "A brief description" {
		t.Errorf("Expected brief 'A brief description', got %q", parsedDoc.FileBrief)
	}
	if parsedDoc.FileDescription != "The full description." {
		t.Errorf("Expected description 'The full description.', got %q", parsedDoc.FileDescription)
	}

	allFuncs := parsedDoc.AllFunctions()
	if len(allFuncs) != 2 {
		t.Fatalf("Expected 2 functions, got %d", len(allFuncs))
	}

	// Check section structure
	if len(parsedDoc.Sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(parsedDoc.Sections))
	}
	if parsedDoc.Sections[0].Name != "Utils" {
		t.Errorf("Expected section name 'Utils', got %q", parsedDoc.Sections[0].Name)
	}
	if parsedDoc.Sections[0].Description != "Helper functions." {
		t.Errorf("Expected section description 'Helper functions.', got %q", parsedDoc.Sections[0].Description)
	}

	// Check first function
	f := allFuncs[0]
	if f.Name != "greet" {
		t.Errorf("Expected function name 'greet', got %q", f.Name)
	}
	if !f.IsDeprecated {
		t.Errorf("Expected IsDeprecated to be true")
	}
	if f.DeprecatedMessage != "Use hello() instead." {
		t.Errorf("Expected DeprecatedMessage 'Use hello() instead.', got %q", f.DeprecatedMessage)
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
	if len(f.Sets) != 1 {
		t.Errorf("Expected 1 set var, got %d", len(f.Sets))
	} else if f.Sets[0].Name != "LAST_GREETED" {
		t.Errorf("Expected set var 'LAST_GREETED', got %q", f.Sets[0].Name)
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
	f2 := allFuncs[1]
	if f2.Name != "farewell" {
		t.Errorf("Expected function name 'farewell', got %q", f2.Name)
	}
	if !f2.IsNoArgs {
		t.Errorf("Expected IsNoArgs to be true")
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

	grouped := loadTestCases(t, true)

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
