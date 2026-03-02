package main

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
)

func TestSchemaValidJSON(t *testing.T) {
	out, err := renderSchema()
	if err != nil {
		t.Fatalf("renderSchema() error: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal([]byte(out), &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	got, ok := schema["$schema"].(string)
	if !ok {
		t.Fatal("$schema field missing or not a string")
	}
	const want = "https://json-schema.org/draft/2020-12/schema"
	if got != want {
		t.Errorf("$schema = %q; want %q", got, want)
	}
}

func TestSchemaTopLevelProperties(t *testing.T) {
	schema := generateSchema()
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("top-level properties missing or wrong type")
	}
	for _, key := range []string{"name", "brief", "description", "functions", "authors", "license", "version"} {
		if _, exists := props[key]; !exists {
			t.Errorf("top-level property %q not found in schema", key)
		}
	}
}

// schemaRequired returns the required list from a schema node.
func schemaRequired(t *testing.T, schema map[string]any, path ...string) []string {
	t.Helper()
	// Navigate to the target node (same logic but return the node itself).
	cur := schema
	for _, seg := range path {
		if seg == "items" {
			raw, ok := cur["items"]
			if !ok {
				t.Fatalf("missing 'items'")
			}
			cur, ok = raw.(map[string]any)
			if !ok {
				t.Fatalf("'items' is not an object")
			}
		} else {
			props, ok := cur["properties"].(map[string]any)
			if !ok {
				t.Fatalf("missing 'properties' before %q", seg)
			}
			raw, ok := props[seg]
			if !ok {
				t.Fatalf("property %q not found", seg)
			}
			cur, ok = raw.(map[string]any)
			if !ok {
				t.Fatalf("property %q not an object", seg)
			}
		}
	}
	raw, ok := cur["required"]
	if !ok {
		return nil
	}
	slice, ok := raw.([]any)
	if !ok {
		t.Fatal("'required' is not an array")
	}
	result := make([]string, len(slice))
	for i, v := range slice {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("required[%d] is not a string", i)
		}
		result[i] = s
	}
	return result
}

func containsAll(haystack []string, needles ...string) bool {
	set := make(map[string]bool, len(haystack))
	for _, s := range haystack {
		set[s] = true
	}
	for _, n := range needles {
		if !set[n] {
			return false
		}
	}
	return true
}

func TestSchemaRequiredFields(t *testing.T) {
	schema := generateSchema()

	tests := []struct {
		name     string
		path     []string
		required []string
	}{
		{
			name:     "JSONFunc",
			path:     []string{"functions", "items"},
			required: []string{"name"},
		},
		{
			name:     "JSONArg",
			path:     []string{"functions", "items", "args", "items"},
			required: []string{"name", "type", "description"},
		},
		{
			name:     "JSONSetVar (set)",
			path:     []string{"functions", "items", "set", "items"},
			required: []string{"name", "type", "description"},
		},
		{
			name:     "JSONExitCode",
			path:     []string{"functions", "items", "exitcodes", "items"},
			required: []string{"code", "description"},
		},
		{
			name:     "OptionEntry",
			path:     []string{"functions", "items", "options", "items"},
			required: []string{"forms", "description"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := schemaRequired(t, schema, tc.path...)
			if !containsAll(req, tc.required...) {
				t.Errorf("required = %v; want to contain %v", req, tc.required)
			}
		})
	}
}

// checkOutputMatchesSchema recursively verifies that every key in jsonObj
// exists in the corresponding schema node's properties.
func checkOutputMatchesSchema(t *testing.T, jsonObj map[string]any, schemaNode map[string]any, path string) {
	t.Helper()
	props, ok := schemaNode["properties"].(map[string]any)
	if !ok {
		return
	}
	for key, val := range jsonObj {
		propSchema, exists := props[key]
		if !exists {
			t.Errorf("JSON key %q (at %s) has no corresponding property in schema", key, path+"."+key)
			continue
		}
		propSchemaMap, ok := propSchema.(map[string]any)
		if !ok {
			continue
		}
		switch v := val.(type) {
		case map[string]any:
			checkOutputMatchesSchema(t, v, propSchemaMap, path+"."+key)
		case []any:
			items, ok := propSchemaMap["items"].(map[string]any)
			if !ok {
				continue
			}
			for i, elem := range v {
				if obj, ok := elem.(map[string]any); ok {
					checkOutputMatchesSchema(t, obj, items, path+"."+key+"[*]")
					_ = i
				}
			}
		}
	}
}

func TestSchemaFileUpToDate(t *testing.T) {
	committed, err := os.ReadFile("schema.json")
	if err != nil {
		t.Fatalf("read schema.json: %v", err)
	}
	generated, err := renderSchema()
	if err != nil {
		t.Fatalf("renderSchema: %v", err)
	}
	if string(committed) != generated {
		t.Error("schema.json is out of date; regenerate with: go run . --schema > schema.json")
	}
}

func TestSchemaMatchesJSONOutput(t *testing.T) {
	f, err := os.Open("examples/showcase.sh")
	if err != nil {
		t.Fatalf("open showcase.sh: %v", err)
	}
	defer f.Close()

	parser := NewParser()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parser.ProcessLine(scanner.Text())
	}

	jsonOut, err := parser.RenderJSON()
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	var jsonDoc map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &jsonDoc); err != nil {
		t.Fatalf("unmarshal JSON output: %v", err)
	}

	schema := generateSchema()
	checkOutputMatchesSchema(t, jsonDoc, schema, "root")
}
