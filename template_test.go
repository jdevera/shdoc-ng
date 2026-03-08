package shdoc

import (
	"os"
	"testing"
)

func TestCustomTemplate(t *testing.T) {
	input := "# @name mylib\n"
	doc, _ := ParseDocument(input)
	out, err := RenderWithTemplate(&doc, `{{.FileTitle}}`)
	if err != nil {
		t.Fatalf("renderWithTemplate error: %v", err)
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
	doc, _ := ParseDocument(string(input))
	out1, err := RenderWithTemplate(&doc, DefaultMarkdownTemplate)
	if err != nil {
		t.Fatalf("renderWithTemplate error: %v", err)
	}
	// Render again with the same template to verify determinism.
	out2, err := RenderWithTemplate(&doc, DefaultMarkdownTemplate)
	if err != nil {
		t.Fatalf("renderWithTemplate (second call) error: %v", err)
	}
	if out1 != out2 {
		t.Errorf("two renders of the same document differ")
	}
}
