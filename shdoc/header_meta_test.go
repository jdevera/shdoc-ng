package shdoc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHeaderMetadataAndChangesParsing(t *testing.T) {
	src := `#!/bin/bash
# @name deploy.sh
# @brief Zero-downtime deployment utility.
# @author Jane Doe <jane@example.com>
# @author Alex Rivera <alex@example.com>
# @version 2.4.0
# @date 2026-11-21
# @license MIT
# @changes 2.4.0 | 2026-11-21 | Added rolling retries and validation.
# @changes 2.3.0 (2026-10-12) Added systemd readiness checks.
# @changelog 2.2.0 2026-09-01 Initial release.
# @example ./deploy.sh --env prod v2.4.0
# @example
#   ./deploy.sh --dry-run v2.4.0
# @description
#     Deployment helper routines.

# @description Deploy a release.
# @arg $1 string Release tag.
deploy_release() {
    echo "$1"
}
`

	doc, warnings := ParseDocument(src)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %+v", warnings)
	}
	if doc.FileTitle != "deploy.sh" {
		t.Fatalf("unexpected FileTitle: %q", doc.FileTitle)
	}
	if doc.Date != "2026-11-21" {
		t.Fatalf("unexpected Date: %q", doc.Date)
	}
	if len(doc.Changes) != 3 {
		t.Fatalf("expected 3 changes, got %d (%+v)", len(doc.Changes), doc.Changes)
	}
	if doc.Changes[0].Version != "2.4.0" || doc.Changes[0].Date != "2026-11-21" {
		t.Fatalf("unexpected first change entry: %+v", doc.Changes[0])
	}
	if doc.Changes[1].Version != "2.3.0" || doc.Changes[1].Date != "2026-10-12" {
		t.Fatalf("unexpected second change entry: %+v", doc.Changes[1])
	}
	if doc.Changes[2].Version != "2.2.0" || doc.Changes[2].Date != "2026-09-01" {
		t.Fatalf("unexpected third change entry: %+v", doc.Changes[2])
	}
	if len(doc.Examples) != 2 {
		t.Fatalf("expected 2 header examples, got %d (%+v)", len(doc.Examples), doc.Examples)
	}
	if strings.TrimSpace(doc.Examples[0]) != "./deploy.sh --env prod v2.4.0" {
		t.Fatalf("unexpected first header example: %q", doc.Examples[0])
	}
	if strings.TrimSpace(doc.Examples[1]) != "./deploy.sh --dry-run v2.4.0" {
		t.Fatalf("unexpected second header example: %q", doc.Examples[1])
	}
}

func TestUnversionedChangesFallback(t *testing.T) {
	src := `#!/bin/bash
# @name script.sh
# @version 1.2.3
# @date 2026-11-20
# @changes Fixed argument quoting.

# @description Ping.
# @noargs
ping() { :; }
`
	doc, warnings := ParseDocument(src)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %+v", warnings)
	}
	if len(doc.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(doc.Changes))
	}
	if doc.Changes[0].Version != "1.2.3" || doc.Changes[0].Date != "2026-11-20" {
		t.Fatalf("expected fallback to @version and @date, got %+v", doc.Changes[0])
	}
}

func TestFileOnlyTagsWarnInsideFunctionBlock(t *testing.T) {
	src := `#!/bin/bash
# @name demo.sh

# @description Function with misplaced tags.
# @date 2026-11-21
# @changes 1.0.0 | Initial
# @noargs
foo() { :; }
`
	_, warnings := ParseDocument(src)
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings for misplaced @date and @changes, got %+v", warnings)
	}
}

func TestMarkdownHTMLAndJSONOutput(t *testing.T) {
	src := `#!/bin/bash
# @name deploy.sh
# @brief Deployment script.
# @author Jane Doe
# @version 2.4.0
# @date 2026-11-21
# @license MIT
# @changes 2.4.0 | 2026-11-21 | Added health retries.
# @example ./deploy.sh --env prod v2.4.0
# @example ./deploy.sh --dry-run v2.4.0

# @description Run deploy.
# @noargs
run_deploy() { :; }
`
	doc, warnings := ParseDocument(src)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", warnings)
	}

	md, err := RenderWithTemplate(&doc, DefaultMarkdownTemplate)
	if err != nil {
		t.Fatalf("RenderWithTemplate markdown failed: %v", err)
	}
	for _, want := range []string{
		"#### Date\n\n2026-11-21",
		"#### What Changed\n\n| Version | Date | Changes |",
		"| `2.4.0` | 2026-11-21 | Added health retries. |",
		"#### Examples",
		"./deploy.sh --env prod v2.4.0",
		"./deploy.sh --dry-run v2.4.0",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q\nOutput:\n%s", want, md)
		}
	}

	htmlOut, err := RenderWithTemplate(&doc, DefaultHTMLTemplate)
	if err != nil {
		t.Fatalf("RenderWithTemplate html failed: %v", err)
	}
	for _, want := range []string{
		`<span class="brand-date">2026-11-21</span>`,
		`<div class="file-doc-meta"`,
		`<table class="data-table changes-table">`,
		`<div class="file-example-list">`,
	} {
		if !strings.Contains(htmlOut, want) {
			t.Errorf("html missing %q", want)
		}
	}

	jsonOut, err := RenderDocumentJSON(&doc)
	if err != nil {
		t.Fatalf("RenderDocumentJSON failed: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if decoded["date"] != "2026-11-21" {
		t.Errorf("json date mismatch: %v", decoded["date"])
	}
}
