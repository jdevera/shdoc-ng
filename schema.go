package main

import (
	"encoding/json"
	"reflect"
	"strings"
)

func parseJSONTag(field reflect.StructField) (name string, omitempty bool, skip bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = field.Name
	}
	for _, p := range parts[1:] {
		if p == "omitempty" {
			omitempty = true
		}
	}
	return name, omitempty, false
}

func schemaForType(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice:
		return map[string]any{
			"type":  "array",
			"items": schemaForType(t.Elem()),
		}
	case reflect.Struct:
		props := map[string]any{}
		var required []any
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			name, omitempty, skip := parseJSONTag(field)
			if skip {
				continue
			}
			props[name] = schemaForType(field.Type)
			if !omitempty {
				required = append(required, name)
			}
		}
		schema := map[string]any{
			"type":                 "object",
			"properties":           props,
			"additionalProperties": false,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	default:
		return map[string]any{}
	}
}

func generateSchema() map[string]any {
	schema := schemaForType(reflect.TypeOf(Document{}))
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	schema["title"] = "shdoc-ng JSON output"
	schema["description"] = "Schema for the JSON output produced by shdoc-ng --format json"
	return schema
}

func renderSchema() (string, error) {
	schema := generateSchema()
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}
