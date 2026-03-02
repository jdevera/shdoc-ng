package main

import (
	"os"
	"strings"
	"testing"
)

func TestCustomTemplate(t *testing.T) {
	input := "# @name mylib\n"
	parser := NewParser()
	for _, line := range strings.Split(input, "\n") {
		parser.ProcessLine(line)
	}
	out, err := parser.RenderWithTemplate(`{{.FileTitle}}`)
	if err != nil {
		t.Fatalf("RenderWithTemplate error: %v", err)
	}
	if out != "mylib" {
		t.Errorf("got %q, want %q", out, "mylib")
	}
}

func TestPrintTemplateRoundtrip(t *testing.T) {
	input, err := os.ReadFile("examples/showcase.sh")
	if err != nil {
		t.Fatalf("open showcase.sh: %v", err)
	}
	parser1 := NewParser()
	parser2 := NewParser()
	for _, line := range strings.Split(string(input), "\n") {
		parser1.ProcessLine(line)
		parser2.ProcessLine(line)
	}
	out1, err := parser1.Render()
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	out2, err := parser2.RenderWithTemplate(defaultMarkdownTemplate)
	if err != nil {
		t.Fatalf("RenderWithTemplate error: %v", err)
	}
	if out1 != out2 {
		t.Errorf("roundtrip output differs from Render()")
	}
}
