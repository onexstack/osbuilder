package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

func applyUpdates(src string, kind string, grpcServiceName string, importPath string) (string, bool, error) {
	changedAny := false

	// 1) Ensure the import exists
	withImport, changed1, err := addImportProto(src, importPath)
	if err != nil {
		return "", false, err
	}
	changedAny = changedAny || changed1

	// 2) Ensure the RPC methods exist
	withRPCs, changed2, err := addRPCsToAPIServer(withImport, kind, grpcServiceName)
	if err != nil {
		return "", false, err
	}
	changedAny = changedAny || changed2

	// Normalize file ending
	withRPCs = normalizeFileEnding(withRPCs)
	return withRPCs, changedAny, nil
}

func backupFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	bak := filepath.Join(dir, base+".bak")
	return os.WriteFile(bak, content, 0o644)
}

// addImportForPostProto inserts the following if missing:
// // 定义当前服务所依赖的博客消息
// import "apiserver/v1/post.proto";
func addImportProto(src string, importPath string) (string, bool, error) {
	var (
		reImportLine       = regexp.MustCompile(`(?m)^[ \t]*import[ \t]+"([^"]+)"[ \t]*;`)
		reTargetImportLine = regexp.MustCompile(fmt.Sprintf(`(?m)^[ \t]*import[ \t]+"%s"[ \t]*;`, importPath))
		rePackageLine      = regexp.MustCompile(`(?m)^[ \t]*package[ \t]+\w+[ \t]*;`)
		reSyntaxLine       = regexp.MustCompile(`(?m)^[ \t]*syntax[ \t]*=[^;]*;`)
	)
	// Already present?
	if reTargetImportLine.FindStringIndex(src) != nil {
		return src, false, nil
	}

	// Determine insertion point and indentation
	insertAt := -1
	indent := ""

	// If there are existing imports, append after the last one and reuse its indentation.
	allImports := reImportLine.FindAllStringIndex(src, -1)
	if len(allImports) > 0 {
		last := allImports[len(allImports)-1]
		insertAt = lineEndIndex(src, last[1])
		indent = lineIndentAt(src, last[0])
	} else {
		// Otherwise, place it after the package line if available; else after syntax.
		if pkg := rePackageLine.FindStringIndex(src); pkg != nil {
			insertAt = lineEndIndex(src, pkg[1])
			indent = lineIndentAt(src, pkg[0])
		} else if syn := reSyntaxLine.FindStringIndex(src); syn != nil {
			insertAt = lineEndIndex(src, syn[1])
			indent = lineIndentAt(src, syn[0])
		} else {
			// Fallback: top of file
			insertAt = 0
			indent = ""
		}
	}

	ins := indent + fmt.Sprintf("import \"%s\";\n", importPath)

	out := src[:insertAt] + ins + src[insertAt:]
	return out, true, nil
}

func addRPCsToAPIServer(src string, kind string, grpcServiceName string) (string, bool, error) {
	var (
		reServiceOpen = regexp.MustCompile(fmt.Sprintf(`(?m)^[ \t]*service[ \t]+%s[ \t]*\{`, grpcServiceName))
		reCreate      = regexp.MustCompile(fmt.Sprintf(`(?m)^[ \t]*rpc[ \t]+Create%s[ \t]*\(`, kind))
		reUpdate      = regexp.MustCompile(fmt.Sprintf(`(?m)^[ \t]*rpc[ \t]+Update%s[ \t]*\(`, kind))
		reDelete      = regexp.MustCompile(fmt.Sprintf(`(?m)^[ \t]*rpc[ \t]+Delete%s[ \t]*\(`, kind))
		reGet         = regexp.MustCompile(fmt.Sprintf(`(?m)^[ \t]*rpc[ \t]+Get%s[ \t]*\(`, kind))
		reList        = regexp.MustCompile(fmt.Sprintf(`(?m)^[ \t]*rpc[ \t]+List%s[ \t]*\(`, kind))
	)

	loc := reServiceOpen.FindStringIndex(src)
	if loc == nil {
		return "", false, errors.New("could not find 'service APIServer {'")
	}

	// Find the opening brace index
	openIdx := strings.Index(src[loc[0]:loc[1]], "{")
	if openIdx < 0 {
		return "", false, errors.New("malformed service declaration: missing '{'")
	}
	openIdx += loc[0]

	closeIdx, err := findMatchingCloseBrace(src, openIdx)
	if err != nil {
		return "", false, err
	}

	// Extract parts
	head := src[:openIdx+1]
	body := src[openIdx+1 : closeIdx]
	tail := src[closeIdx:] // includes the closing '}' and after

	// Check existing
	hasCreate := reCreate.FindStringIndex(body) != nil
	hasUpdate := reUpdate.FindStringIndex(body) != nil
	hasDelete := reDelete.FindStringIndex(body) != nil
	hasGet := reGet.FindStringIndex(body) != nil
	hasList := reList.FindStringIndex(body) != nil

	insIndent := inferIndent(body)

	var b strings.Builder
	b.WriteString(head)
	// Ensure body keeps as-is, but we may need a newline before insertion
	bodyTrimRight := strings.TrimRight(body, " \t")
	b.WriteString(bodyTrimRight)

	needsTrailingNewline := !strings.HasSuffix(bodyTrimRight, "\n")
	if needsTrailingNewline {
		// b.WriteString("\n")
	}

	// Prepare the lines to insert, only missing ones
	needUpdate := false
	// Prepare the lines to insert, only missing ones
	if !hasCreate {
		b.WriteString(insIndent)
		b.WriteString(fmt.Sprintf("rpc Create%s(Create%sRequest) returns (Create%sResponse);\n", kind, kind, kind))
		needUpdate = true
	}
	if !hasUpdate {
		b.WriteString(insIndent)
		b.WriteString(fmt.Sprintf("rpc Update%s(Update%sRequest) returns (Update%sResponse);\n", kind, kind, kind))
	}
	if !hasDelete {
		b.WriteString(insIndent)
		b.WriteString(fmt.Sprintf("rpc Delete%s(Delete%sRequest) returns (Delete%sResponse);\n", kind, kind, kind))
		needUpdate = true
	}
	if !hasGet {
		b.WriteString(insIndent)
		b.WriteString(fmt.Sprintf("rpc Get%s(Get%sRequest) returns (Get%sResponse);\n", kind, kind, kind))
		needUpdate = true
	}
	if !hasList {
		b.WriteString(insIndent)
		b.WriteString(fmt.Sprintf("rpc List%s(List%sRequest) returns (List%sResponse);\n", kind, kind, kind))
		needUpdate = true
	}

	if !needUpdate {
		return src, false, nil
	}

	// Preserve original trailing whitespace in body
	b.WriteString(trailingWhitespace(body))

	// Append closing brace + rest
	b.WriteString(tail)

	out := normalizeFileEnding(b.String())
	return out, true, nil
}

func inferIndent(body string) string {
	// Default two spaces
	def := "  "
	lines := strings.Split(body, "\n")
	minSpaces := -1
	var candidate string
	for _, ln := range lines {
		l := strings.TrimRight(ln, " \t\r")
		if strings.TrimSpace(l) == "" {
			continue
		}
		lead := leadingWhitespace(l)
		// We only consider lines that are indented at least one space or tab
		if len(lead) == 0 {
			continue
		}
		// prefer the smallest non-zero indentation
		spaces := visualWidth(lead)
		if minSpaces == -1 || spaces < minSpaces {
			minSpaces = spaces
			candidate = lead
		}
	}
	if candidate != "" {
		return candidate
	}
	return def
}

func leadingWhitespace(s string) string {
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			break
		}
	}
	return s[:i]
}

func trailingWhitespace(s string) string {
	i := len(s) - 1
	for i >= 0 {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\r' && s[i] != '\n' {
			break
		}
		i--
	}
	return s[i+1:]
}

func visualWidth(ws string) int {
	// Treat tab as 4 spaces for width approximation
	width := 0
	for len(ws) > 0 {
		r, size := utf8.DecodeRuneInString(ws)
		ws = ws[size:]
		if r == '\t' {
			width += 4
		} else if r == ' ' {
			width++
		}
	}
	return width
}

func findMatchingCloseBrace(s string, open int) (int, error) {
	depth := 0
	inLineComment := false
	inBlockComment := false
	inString := false
	var stringQuote rune

	for i := open; i < len(s); {
		if inLineComment {
			if s[i] == '\n' {
				inLineComment = false
			}
			i++
			continue
		}
		if inBlockComment {
			if i+1 < len(s) && s[i] == '*' && s[i+1] == '/' {
				inBlockComment = false
				i += 2
				continue
			}
			i++
			continue
		}
		if inString {
			r, sz := utf8.DecodeRuneInString(s[i:])
			i += sz
			if r == '\\' { // escape next
				_, sz2 := utf8.DecodeRuneInString(s[i:])
				i += sz2
				continue
			}
			if r == stringQuote {
				inString = false
			}
			continue
		}

		// Not in any comment/string: check for start of comment or string
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '/' {
			inLineComment = true
			i += 2
			continue
		}
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '*' {
			inBlockComment = true
			i += 2
			continue
		}
		if s[i] == '"' || s[i] == '\'' {
			stringQuote = rune(s[i])
			inString = true
			i++
			continue
		}

		// Braces
		if s[i] == '{' {
			depth++
			i++
			continue
		}
		if s[i] == '}' {
			depth--
			if depth == 0 {
				return i, nil
			}
			i++
			continue
		}
		i++
	}
	return 0, errors.New("unbalanced braces when parsing service APIServer block")
}

func normalizeFileEnding(s string) string {
	// Ensure the file ends with a single newline
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if !strings.HasSuffix(s, "\n") {
		return s + "\n"
	}
	// Reduce multiple trailing newlines to one
	return strings.TrimRightFunc(s, func(r rune) bool { return r == '\n' }) + "\n"
}

func info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintln(os.Stderr, "[info]", msg)
}

func fatal(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintln(os.Stderr, "[error]", msg)
	os.Exit(1)
}

// Helpers for line/indent

func lineEndIndex(s string, at int) int {
	// Return index right AFTER the newline terminating the line that contains position 'at-1'
	if at < 0 {
		return 0
	}
	n := strings.IndexByte(s[at:], '\n')
	if n == -1 {
		return len(s)
	}
	return at + n + 1
}

func lineStartIndex(s string, at int) int {
	if at <= 0 {
		return 0
	}
	i := at - 1
	for i >= 0 && s[i] != '\n' {
		i--
	}
	return i + 1
}

func lineIndentAt(s string, start int) string {
	ls := lineStartIndex(s, start)
	end := ls
	for end < len(s) && (s[end] == ' ' || s[end] == '\t') {
		end++
	}
	return s[ls:end]
}
