package render

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lxa-project/lxa/internal/scanner"
)

// Renderer handles formatting output.
type Renderer struct {
	out             io.Writer
	opts            Options
	termWidth       int
	isTerminal      bool
	jsonOutputFound bool
}

// Options configure the renderer.
type Options struct {
	JSONOutput      bool
	Inspect         bool
	MaxTagsWidth    int
	MaxCommentWidth int
	NoWrap          bool
}

// New creates a new Renderer.
func New(out io.Writer, opts Options) *Renderer {
	r := &Renderer{
		out:  out,
		opts: opts,
	}

	if f, ok := out.(*os.File); ok {
		r.isTerminal = IsTerminal(f)
		r.termWidth = TerminalWidth(f)
	}

	if r.opts.JSONOutput {
		fmt.Fprint(r.out, "[")
	}

	return r
}

// File handles an incoming FileInfo.
func (r *Renderer) File(f scanner.FileInfo) {
	if r.opts.JSONOutput {
		if r.jsonOutputFound {
			fmt.Fprint(r.out, ",\n")
		} else {
			fmt.Fprint(r.out, "\n")
		}
		r.jsonOutputFound = true

		// Construct JSON representation
		v := map[string]interface{}{
			"path": f.Path,
		}
		if f.Error != nil {
			v["error"] = f.Error.Error()
		} else {
			v["type"] = f.Info.Mode().String()
			v["size"] = f.Info.Size()
			v["metadata"] = f.Metadata
		}

		b, _ := json.MarshalIndent(v, "  ", "  ")
		fmt.Fprint(r.out, string(b))
		return
	}

	if r.opts.Inspect {
		r.renderInspect(f)
		return
	}

	r.renderList(f)
}

func (r *Renderer) renderList(f scanner.FileInfo) {
	if f.Error != nil {
		fmt.Fprintf(r.out, "%s (error: %s)\n", f.Path, f.Error)
		return
	}

	name := f.Path

	var parts []string
	if f.Metadata.HasTags {
		tagsStr := strings.Join(f.Metadata.Tags, ",")
		tagsStr = r.truncate(tagsStr, r.opts.MaxTagsWidth)
		parts = append(parts, fmt.Sprintf("[tags: %s]", tagsStr))
	}
	if f.Metadata.HasCmnt {
		cmntStr := f.Metadata.Comment
		cmntStr = r.truncate(cmntStr, r.opts.MaxCommentWidth)
		parts = append(parts, fmt.Sprintf("[comment: %s]", cmntStr))
	}

	var metaStr string
	if len(parts) > 0 {
		metaStr = " " + strings.Join(parts, " ")
	}

	line := name + metaStr
	fmt.Fprintln(r.out, line)
}

func (r *Renderer) renderInspect(f scanner.FileInfo) {
	fmt.Fprintf(r.out, "Path: %s\n", f.Path)
	if f.Error != nil {
		fmt.Fprintf(r.out, "  Error: %s\n\n", f.Error)
		return
	}

	fmt.Fprintf(r.out, "  Type: %s\n", f.Info.Mode())
	fmt.Fprintf(r.out, "  Size: %d\n", f.Info.Size())

	if !f.Metadata.HasXDG {
		fmt.Fprintln(r.out, "  XDG Metadata: (none)")
	} else {
		fmt.Fprintln(r.out, "  XDG Metadata:")
		if f.Metadata.HasTags {
			fmt.Fprintf(r.out, "    tags: %s\n", strings.Join(f.Metadata.Tags, ", "))
		}
		if f.Metadata.HasCmnt {
			fmt.Fprintf(r.out, "    comment: %s\n", f.Metadata.Comment)
		}
		for k, v := range f.Metadata.XDG {
			if k != "user.xdg.tags" && k != "user.xdg.comment" {
				fmt.Fprintf(r.out, "    %s: %s\n", strings.TrimPrefix(k, "user.xdg."), string(v))
			}
		}
	}

	var hasOther bool
	for k := range f.Metadata.All {
		if !strings.HasPrefix(k, "user.xdg.") {
			hasOther = true
			break
		}
	}

	if hasOther {
		fmt.Fprintln(r.out, "  Other xattrs:")
		for k, v := range f.Metadata.All {
			if !strings.HasPrefix(k, "user.xdg.") {
				fmt.Fprintf(r.out, "    %s: %s\n", k, string(v))
			}
		}
	}
	fmt.Fprintln(r.out)
}

// Close finishes the output.
func (r *Renderer) Close() {
	if r.opts.JSONOutput {
		fmt.Fprintln(r.out, "\n]")
	}
}

func (r *Renderer) truncate(s string, max int) string {
	if !r.opts.NoWrap && r.isTerminal && r.termWidth > 0 {
		if r.termWidth < max {
			max = r.termWidth / 3
		}
	}
	if max <= 0 {
		return s
	}

	// Calculate a dynamic max if terminal width is known and not disabled
	if !r.opts.NoWrap && r.isTerminal && r.termWidth > 0 {
		// Just use `max` as the ceiling, we could make it smarter to fill terminal
	}

	// simple utf8 aware truncate
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max-1]) + "…" // U+2026 HORIZONTAL ELLIPSIS
	}
	return s
}
