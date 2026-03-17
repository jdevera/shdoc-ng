package shdoc

import "regexp"

// CommentBlock is a run of consecutive comment lines.
type CommentBlock struct {
	Lines    []LexedLine
	StartNum int
	EndNum   int // Num of the last line in Lines
}

// BlockKind distinguishes function doc blocks from file-level meta blocks.
type BlockKind int

const (
	FuncDocBlockKind BlockKind = iota
	MetaBlockKind
)

// ParsedBlock is a CommentBlock paired with its context.
type ParsedBlock struct {
	Kind     BlockKind
	Comments CommentBlock
	FuncName string // non-empty when Kind == FuncDocBlockKind
}

var (
	segFuncDeclWithBrace = regexp.MustCompile(
		`^[\t ]*(function[\t ]+)?([a-zA-Z0-9_\-:.]+)[\t ]*(\([\t ]*\))?[\t ]*[{(]`,
	)
	segFuncDeclWithoutBrace = regexp.MustCompile(
		`^[\t ]*(function[\t ]+)?([a-zA-Z0-9_\-:.]+)[\t ]*(\([\t ]*\))?[\t ]*$`,
	)
	segLoneBrace  = regexp.MustCompile(`^[\t ]*[{(]`)
	segFuncNameRe = regexp.MustCompile(
		`^\s*(?:function\s+)?([a-zA-Z0-9_\-:.]+)\s*(?:\(\s*\))?\s*\{?`,
	)
)

// segmentBlocks walks lexed lines and groups consecutive comment lines into
// CommentBlocks, pairing each with the function declaration that follows it
// (if any).
func SegmentBlocks(lines []LexedLine) []ParsedBlock {
	var blocks []ParsedBlock
	i, n := 0, len(lines)

	for i < n {
		if lines[i].Kind != LineComment {
			i++
			continue
		}

		// Collect consecutive comment lines.
		start := i
		for i < n && lines[i].Kind == LineComment {
			i++
		}
		block := CommentBlock{
			Lines:    lines[start:i],
			StartNum: lines[start].Num,
			EndNum:   lines[i-1].Num,
		}

		// Look for an immediately-following function declaration.
		funcName := ""
		if i < n && lines[i].Kind == LineCode {
			raw := lines[i].Raw
			if segFuncDeclWithBrace.MatchString(raw) {
				funcName = ExtractFuncName(raw)
				i++ // consume the declaration line
			} else if segFuncDeclWithoutBrace.MatchString(raw) {
				declLine := raw
				i++ // consume the declaration line
				// The brace may be on the very next line, with at most one
				// blank-or-code line between declaration and brace.
				if i < n && (lines[i].Kind == LineBlank || lines[i].Kind == LineCode) {
					if segLoneBrace.MatchString(lines[i].Raw) {
						funcName = ExtractFuncName(declLine)
						i++ // consume the brace line
					}
				}
			}
		}

		if funcName != "" {
			blocks = append(blocks, ParsedBlock{
				Kind:     FuncDocBlockKind,
				Comments: block,
				FuncName: funcName,
			})
		} else {
			blocks = append(blocks, ParsedBlock{
				Kind:     MetaBlockKind,
				Comments: block,
			})
		}
	}

	return blocks
}

// IsFuncDecl reports whether line looks like a shell function declaration.
func IsFuncDecl(line string) bool {
	return segFuncDeclWithBrace.MatchString(line)
}

// ExtractFuncName pulls the function name from a declaration line.
func ExtractFuncName(line string) string {
	if m := segFuncNameRe.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	return ""
}
