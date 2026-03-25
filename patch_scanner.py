import re

with open("internal/scanner/scanner.go", "r") as f:
    content = f.read()

opts_search = """type Options struct {
	Recursive bool
	Filter    string
	XDGOnly   bool // shorthand for filter "xdg"
	FS        FileSystem
}"""
opts_replace = """type Options struct {
	Recursive  bool
	Filter     string
	XDGOnly    bool // shorthand for filter "xdg"
	FS         FileSystem
	ShowHidden bool
}"""
content = content.replace(opts_search, opts_replace)

emit_search = """	if s.opts.XDGOnly && !md.HasXDG {
		return
	}"""
emit_replace = """	if !s.opts.ShowHidden && strings.HasPrefix(filepath.Base(path), ".") && path != "." && path != ".." && path != root {
		return
	}

	if s.opts.XDGOnly && !md.HasXDG {
		return
	}"""
content = content.replace(emit_search, emit_replace)

import_search = """	"github.com/lxa-project/lxa/internal/filter"
	"github.com/lxa-project/lxa/internal/xattr"
)"""
import_replace = """	"strings"

	"github.com/lxa-project/lxa/internal/filter"
	"github.com/lxa-project/lxa/internal/xattr"
)"""
content = content.replace(import_search, import_replace)

with open("internal/scanner/scanner.go", "w") as f:
    f.write(content)
