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
	columnsOutput   []string // buffer for multi-column output
}

// Options configure the renderer.
type Options struct {
	JSONOutput      bool
	Inspect         bool
	MaxTagsWidth    int
	MaxCommentWidth int
	NoWrap          bool
	NoHeader        bool
	LongListing     bool
	NoGroup         bool
	NoUser          bool
	ShowHeader      bool // changed from ShowTitle
	ShowAuthor      bool
	ShowCreator     bool
	ShowOrigin      bool
	ShowChecksum    bool
	SingleColumn    bool // -1 flag
	MultiColumn     bool // -C flag
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
	} else if !r.opts.Inspect && !r.opts.MultiColumn {
		r.tw = tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	}
	r.columnsOutput = []string{}

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

	if r.opts.ShowHeader && !r.opts.NoHeader && !r.headerPrinted && !r.opts.MultiColumn {
		header := []string{}
		if r.opts.LongListing {
			header = append(header, "PERMISSIONS", "NODE")
			if !r.opts.NoUser {
				header = append(header, "OWNER")
			}
			if !r.opts.NoGroup {
				header = append(header, "GROUP")
			}
			header = append(header, "SIZE", "MODIFIED")
		}

		header = append(header, "FILENAME")

		if r.opts.ShowAuthor {
			header = append(header, "AUTHOR")
		}
		if r.opts.ShowCreator {
			header = append(header, "CREATOR")
		}
		if r.opts.ShowOrigin {
			header = append(header, "ORIGIN")
		}
		if r.opts.ShowChecksum {
			header = append(header, "CHECKSUM")
		}

		header = append(header, "TAGS", "COMMENTS")

		_, _ = fmt.Fprintln(r.tw, strings.Join(header, "\t"))
		r.headerPrinted = true
	}

	r.renderList(f)
}

// formatSize converts bytes to human readable format
func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatTime converts time to match standard ls -l like "Jan  1  2024" or "Jan  1 15:04"
func formatTime(t time.Time) string {
	now := time.Now()
	year := t.Year()
	if year == now.Year() {
		return t.Format("Jan _2 15:04")
	}
	return t.Format("Jan _2  2006")
}

func (r *Renderer) renderList(f scanner.FileInfo) {
	if f.Error != nil {
		_, _ = fmt.Fprintf(r.tw, "-\t-\t-\t-\t-\t-\t%s\t(error: %s)\t\n", f.Path, f.Error)
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
	size := formatSize(f.Info.Size())
	modTime := formatTime(f.Info.ModTime())

	owner := "-"
	group := "-"
	node := "1"

	if stat, ok := f.Info.Sys().(*syscall.Stat_t); ok {
		node = fmt.Sprint(stat.Nlink)

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

	cols := []string{}
	if r.opts.LongListing {
		cols = append(cols, mode, node)
		if !r.opts.NoUser {
			cols = append(cols, owner)
		}
		if !r.opts.NoGroup {
			cols = append(cols, group)
		}
		cols = append(cols, size, modTime)
	}

	cols = append(cols, name)

	if r.opts.ShowAuthor {
		author := ""
		if v, ok := f.Metadata.All["user.author"]; ok {
			author = string(v)
		}
		cols = append(cols, r.truncate(author, 20))
	}
	if r.opts.ShowCreator {
		creator := ""
		if v, ok := f.Metadata.All["user.creator"]; ok {
			creator = string(v)
		}
		cols = append(cols, r.truncate(creator, 20))
	}
	if r.opts.ShowOrigin {
		origin := ""
		if v, ok := f.Metadata.All["user.origin"]; ok {
			origin = string(v)
		}
		cols = append(cols, r.truncate(origin, 40))
	}
	if r.opts.ShowChecksum {
		checksum := ""
		if v, ok := f.Metadata.All["user.checksum"]; ok {
			checksum = string(v)
		}
		cols = append(cols, r.truncate(checksum, 32))
	}

	cols = append(cols, tagsStr, cmntStr)

	if r.opts.MultiColumn {
		r.columnsOutput = append(r.columnsOutput, strings.Join(cols, "\t"))
	} else {
		_, _ = fmt.Fprintln(r.tw, strings.Join(cols, "\t"))
	}
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
	} else if r.opts.MultiColumn {
		// Calculate columns layout
		if len(r.columnsOutput) == 0 {
			return
		}

		// For a simple column layout, we use a tabwriter but we write entries horizontally
		// To match `ls -C`, we format items side by side
		colTw := tabwriter.NewWriter(r.out, 0, 0, 4, ' ', 0)

		// Very basic horizontal wrapping based on terminal width (if available)
		// Assuming termWidth ~80 if not terminal
		width := r.termWidth
		if width <= 0 {
			width = 80
		}

		var line []string
		var currWidth int

		for _, item := range r.columnsOutput {
			itemLen := len(item) + 4 // roughly adding tab width
			if currWidth+itemLen > width && len(line) > 0 {
				_, _ = fmt.Fprintln(colTw, strings.Join(line, "\t"))
				line = []string{item}
				currWidth = itemLen
			} else {
				line = append(line, item)
				currWidth += itemLen
			}
		}
		if len(line) > 0 {
			_, _ = fmt.Fprintln(colTw, strings.Join(line, "\t"))
		}
		colTw.Flush()
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
