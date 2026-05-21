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
			if lines[i].Kind == LineCode {
				if name, consumed := matchBareFuncDecl(lines, i); consumed > 0 {
					blocks = append(blocks, ParsedBlock{
						Kind:     FuncDocBlockKind,
						FuncName: name,
					})
					i += consumed
					continue
				}
			}
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

// shellReservedWords are shell keywords that can appear at the start of a
// compound command and superficially match the function-declaration regex
// (e.g. `for ((...))`, `while ((...))`). They can never be function names.
var shellReservedWords = map[string]bool{
	"if": true, "then": true, "else": true, "elif": true, "fi": true,
	"case": true, "esac": true,
	"for": true, "while": true, "until": true, "do": true, "done": true,
	"select": true, "function": true, "in": true, "time": true,
}

// IsFuncDecl reports whether line looks like a shell function declaration.
func IsFuncDecl(line string) bool {
	return segFuncDeclWithBrace.MatchString(line) && ExtractFuncName(line) != ""
}

// matchBareFuncDecl checks whether lines[i] (assumed LineCode) is the start of
// a function declaration, using the same recognition rules SegmentBlocks
// applies to a function that follows a comment block. Returns the function
// name and the number of lines consumed (1 or 2). Returns "", 0 if no match.
func matchBareFuncDecl(lines []LexedLine, i int) (string, int) {
	raw := lines[i].Raw
	if segFuncDeclWithBrace.MatchString(raw) {
		if name := ExtractFuncName(raw); name != "" {
			return name, 1
		}
	}
	if segFuncDeclWithoutBrace.MatchString(raw) {
		if i+1 < len(lines) && (lines[i+1].Kind == LineBlank || lines[i+1].Kind == LineCode) {
			if segLoneBrace.MatchString(lines[i+1].Raw) {
				if name := ExtractFuncName(raw); name != "" {
					return name, 2
				}
			}
		}
	}
	return "", 0
}

// ExtractFuncName pulls the function name from a declaration line.
// Returns "" if the matched identifier is a shell reserved word (e.g. `for`
// in `for ((i=0; i<10; i++))`).
func ExtractFuncName(line string) string {
	m := segFuncNameRe.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	if shellReservedWords[m[1]] {
		return ""
	}
	return m[1]
}
