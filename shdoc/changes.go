package shdoc

import (
	"regexp"
	"strings"
)

var (
	changeSemverTokenRe   = regexp.MustCompile(`^v?\d+\.\d+(?:\.\d+)?(?:-[a-zA-Z0-9.-]+)?$`)
	changeDateTokenRe     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	changeParenDateRe     = regexp.MustCompile(`^\((\d{4}-\d{2}-\d{2})\)\s+(.*)$`)
	changeDatePrefixRe    = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\s+(.*)$`)
	changeWhitespaceSplit = regexp.MustCompile(`^(\S+)\s+(.*)$`)
)

// parseVersionedChangeLine parses a single @changes line in one of these forms:
//   - "2.4.0 | 2026-11-21 | Added retries"
//   - "2.4.0 | Added retries"
//   - "2.4.0 (2026-11-21) Added retries"
//   - "2.4.0 2026-11-21 Added retries"
//   - "2.4.0 Added retries"
func parseVersionedChangeLine(text string) (ChangeEntry, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ChangeEntry{}, false
	}

	if strings.Contains(trimmed, "|") {
		rawParts := strings.Split(trimmed, "|")
		parts := make([]string, len(rawParts))
		for i, p := range rawParts {
			parts[i] = strings.TrimSpace(p)
		}
		if len(parts) >= 3 && parts[0] != "" {
			version := parts[0]
			maybeDate := parts[1]
			desc := strings.TrimSpace(strings.Join(parts[2:], " | "))
			if desc != "" {
				return ChangeEntry{
					Version:     version,
					Date:        maybeDate,
					Description: desc,
				}, true
			}
		} else if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return ChangeEntry{
				Version:     parts[0],
				Description: parts[1],
			}, true
		}
	}

	if m := changeWhitespaceSplit.FindStringSubmatch(trimmed); m != nil && changeSemverTokenRe.MatchString(m[1]) {
		version := m[1]
		rest := strings.TrimSpace(m[2])
		if pm := changeParenDateRe.FindStringSubmatch(rest); pm != nil {
			return ChangeEntry{
				Version:     version,
				Date:        pm[1],
				Description: strings.TrimSpace(pm[2]),
			}, true
		}
		if dm := changeDatePrefixRe.FindStringSubmatch(rest); dm != nil && changeDateTokenRe.MatchString(dm[1]) {
			return ChangeEntry{
				Version:     version,
				Date:        dm[1],
				Description: strings.TrimSpace(dm[2]),
			}, true
		}
		if rest != "" {
			return ChangeEntry{
				Version:     version,
				Description: rest,
			}, true
		}
	}

	return ChangeEntry{}, false
}

// finalizeDocumentChanges normalizes unversioned @changes rows by inheriting
// document-level Version and Date when present.
func finalizeDocumentChanges(doc *Document) {
	if doc == nil || len(doc.Changes) == 0 {
		return
	}
	for i := range doc.Changes {
		if doc.Changes[i].Version == "" {
			if doc.Version != "" {
				doc.Changes[i].Version = doc.Version
			} else {
				doc.Changes[i].Version = "—"
			}
			if doc.Changes[i].Date == "" && doc.Date != "" {
				doc.Changes[i].Date = doc.Date
			}
		}
	}
}
