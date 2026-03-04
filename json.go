package main

import (
	"bytes"
	"encoding/json"
)

// renderDocumentJSON produces JSON output for a Document.
func renderDocumentJSON(doc *Document) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}
