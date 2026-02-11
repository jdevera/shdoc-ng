package main

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
