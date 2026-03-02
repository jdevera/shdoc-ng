package main

import (
	"bytes"
	"encoding/json"
)

// RenderJSON produces JSON output for the parsed document.
func (p *Parser) RenderJSON() (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&p.doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}
