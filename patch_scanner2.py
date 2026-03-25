import re

with open("internal/scanner/scanner.go", "r") as f:
    content = f.read()

emit_search = """func (s *Scanner) emitIfMatches(path string, info fs.FileInfo, out chan<- FileInfo) {
	if !s.opts.ShowHidden && strings.HasPrefix(filepath.Base(path), ".") && path != "." && path != ".." && path != root {
		return
	}"""
emit_replace = """func (s *Scanner) emitIfMatches(path string, info fs.FileInfo, out chan<- FileInfo) {
	if !s.opts.ShowHidden && strings.HasPrefix(filepath.Base(path), ".") && path != "." && path != ".." {
		return
	}"""

content = content.replace(emit_search, emit_replace)

with open("internal/scanner/scanner.go", "w") as f:
    f.write(content)
