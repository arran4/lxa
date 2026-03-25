package render

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/lxa-project/lxa/internal/scanner"
)

// Renderer handles formatting output.
type Renderer struct {
	out             io.Writer
	tw              *tabwriter.Writer
	opts            Options
	termWidth       int
	isTerminal      bool
	jsonOutputFound bool
	headerPrinted   bool
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
		_, _ = fmt.Fprint(r.out, "[")
	} else if !r.opts.Inspect {
		r.tw = tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	}

	return r
}

// File handles an incoming FileInfo.
func (r *Renderer) File(f scanner.FileInfo) {
	if r.opts.JSONOutput {
		if r.jsonOutputFound {
			_, _ = fmt.Fprint(r.out, ",\n")
		} else {
			_, _ = fmt.Fprint(r.out, "\n")
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
		_, _ = fmt.Fprint(r.out, string(b))
		return
	}

	if r.opts.Inspect {
		r.renderInspect(f)
		return
	}

	if !r.headerPrinted {
		_, _ = fmt.Fprintln(r.tw, "FILENAME\tPERMISSIONS\tOWNER\tGROUP\tSIZE\tMODIFIED\tTAGS\tCOMMENTS")
		r.headerPrinted = true
	}

	r.renderList(f)
}

func (r *Renderer) renderList(f scanner.FileInfo) {
	if f.Error != nil {
		_, _ = fmt.Fprintf(r.tw, "%s\t-\t-\t-\t-\t-\t(error: %s)\t\n", f.Path, f.Error)
		return
	}

	name := f.Path

	tagsStr := ""
	if f.Metadata.HasTags {
		tagsStr = strings.Join(f.Metadata.Tags, ",")
		tagsStr = r.truncate(tagsStr, r.opts.MaxTagsWidth)
	}

	cmntStr := ""
	if f.Metadata.HasCmnt {
		cmntStr = f.Metadata.Comment
		cmntStr = r.truncate(cmntStr, r.opts.MaxCommentWidth)
	}

	// File info extraction
	mode := f.Info.Mode().String()
	size := fmt.Sprintf("%d", f.Info.Size())
	modTime := f.Info.ModTime().Format(time.Stamp)

	owner := "-"
	group := "-"
	if stat, ok := f.Info.Sys().(*syscall.Stat_t); ok {
		if u, err := user.LookupId(fmt.Sprint(stat.Uid)); err == nil {
			owner = u.Username
		} else {
			owner = fmt.Sprint(stat.Uid)
		}
		if g, err := user.LookupGroupId(fmt.Sprint(stat.Gid)); err == nil {
			group = g.Name
		} else {
			group = fmt.Sprint(stat.Gid)
		}
	}

	_, _ = fmt.Fprintf(r.tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", name, mode, owner, group, size, modTime, tagsStr, cmntStr)
}

func (r *Renderer) renderInspect(f scanner.FileInfo) {
	_, _ = fmt.Fprintf(r.out, "Path: %s\n", f.Path)
	if f.Error != nil {
		_, _ = fmt.Fprintf(r.out, "  Error: %s\n\n", f.Error)
		return
	}

	_, _ = fmt.Fprintf(r.out, "  Type: %s\n", f.Info.Mode())
	_, _ = fmt.Fprintf(r.out, "  Size: %d\n", f.Info.Size())

	if !f.Metadata.HasXDG {
		_, _ = fmt.Fprintln(r.out, "  XDG Metadata: (none)")
	} else {
		_, _ = fmt.Fprintln(r.out, "  XDG Metadata:")
		if f.Metadata.HasTags {
			_, _ = fmt.Fprintf(r.out, "    tags: %s\n", strings.Join(f.Metadata.Tags, ", "))
		}
		if f.Metadata.HasCmnt {
			_, _ = fmt.Fprintf(r.out, "    comment: %s\n", f.Metadata.Comment)
		}
		for k, v := range f.Metadata.XDG {
			if k != "user.xdg.tags" && k != "user.xdg.comment" {
				_, _ = fmt.Fprintf(r.out, "    %s: %s\n", strings.TrimPrefix(k, "user.xdg."), string(v))
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
		_, _ = fmt.Fprintln(r.out, "  Other xattrs:")
		for k, v := range f.Metadata.All {
			if !strings.HasPrefix(k, "user.xdg.") {
				_, _ = fmt.Fprintf(r.out, "    %s: %s\n", k, string(v))
			}
		}
	}
	_, _ = fmt.Fprintln(r.out)
}

// Close finishes the output.
func (r *Renderer) Close() {
	if r.opts.JSONOutput {
		_, _ = fmt.Fprintln(r.out, "\n]")
	} else if r.tw != nil {
		r.tw.Flush()
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

	// simple utf8 aware truncate
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max-1]) + "…" // U+2026 HORIZONTAL ELLIPSIS
	}
	return s
}
