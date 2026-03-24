package cli

import (
	"flag"
	"fmt"
	"io"
	"sort"

	"github.com/lxa-project/lxa/internal/render"
	"github.com/lxa-project/lxa/internal/scanner"
	"github.com/lxa-project/lxa/internal/xattr"
)

// Run executes the application.
func Run(args []string, out io.Writer, errOut io.Writer) int {
	f := flag.NewFlagSet("lxa", flag.ContinueOnError)
	f.SetOutput(errOut)

	var (
		xdgOnly    = f.Bool("xdg-only", false, "Show only files with XDG metadata")
		tagsOnly   = f.Bool("tags", false, "Show only files with XDG tags")
		cmntOnly   = f.Bool("comments", false, "Show only files with an XDG comment")
		allXDG     = f.Bool("all-xdg", false, "Show all XDG metadata attributes") // reserved for inspect mode currently
		allXattr   = f.Bool("all-xattr", false, "Show all xattrs")                // reserved for inspect mode
		jsonOutput = f.Bool("json", false, "Output in JSON format")
		recursive  = f.Bool("recursive", false, "Traverse directories recursively")
		filterExpr = f.String("filter", "", "Apply filter expression (e.g. 'has:tags && !has:comment')")
		maxTagsW   = f.Int("max-tags-width", 40, "Maximum display width for tags (0 to disable)")
		maxCmntW   = f.Int("max-comment-width", 60, "Maximum display width for comments (0 to disable)")
		sortField  = f.String("sort", "name", "Sort by: name, path, xdg, tags, comment")
	)

	f.Usage = func() {
		fmt.Fprintln(errOut, "Usage: lxa [OPTIONS] [PATH...]")
		fmt.Fprintln(errOut, "       lxa inspect [PATH...]")
		fmt.Fprintln(errOut, "\nlxa is a Linux-first file listing tool focused on extended attributes and XDG metadata.")
		fmt.Fprintln(errOut, "\nOptions:")
		f.PrintDefaults()
	}

	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	cmdArgs := f.Args()

	inspectMode := false
	if len(cmdArgs) > 0 && cmdArgs[0] == "inspect" {
		inspectMode = true
		cmdArgs = cmdArgs[1:]
	}

	if len(cmdArgs) == 0 {
		cmdArgs = []string{"."}
	}

	// Compile filters from shorthand flags
	finalFilter := *filterExpr
	if *tagsOnly {
		if finalFilter != "" {
			finalFilter = "(" + finalFilter + ") && has:tags"
		} else {
			finalFilter = "has:tags"
		}
	}
	if *cmntOnly {
		if finalFilter != "" {
			finalFilter = "(" + finalFilter + ") && has:comment"
		} else {
			finalFilter = "has:comment"
		}
	}

	scanOpts := scanner.Options{
		Recursive: *recursive,
		Filter:    finalFilter,
		XDGOnly:   *xdgOnly,
	}

	scan, err := scanner.New(xattr.NewSyscallReader(), scanOpts)
	if err != nil {
		fmt.Fprintf(errOut, "Error initializing scanner: %v\n", err)
		return 1
	}

	renderOpts := render.Options{
		JSONOutput:      *jsonOutput,
		Inspect:         inspectMode || *allXDG || *allXattr,
		MaxTagsWidth:    *maxTagsW,
		MaxCommentWidth: *maxCmntW,
		NoWrap:          false,
	}

	ch := scan.Scan(cmdArgs)
	var files []scanner.FileInfo
	for fileInfo := range ch {
		files = append(files, fileInfo)
	}

	// Sorting
	sort.Slice(files, func(i, j int) bool {
		f1, f2 := files[i], files[j]

		switch *sortField {
		case "path":
			return f1.Path < f2.Path
		case "xdg":
			if f1.Metadata.HasXDG != f2.Metadata.HasXDG {
				return f1.Metadata.HasXDG // files with XDG first
			}
			return f1.Path < f2.Path
		case "tags":
			if f1.Metadata.HasTags != f2.Metadata.HasTags {
				return f1.Metadata.HasTags // files with tags first
			}
			return f1.Path < f2.Path
		case "comment":
			if f1.Metadata.HasCmnt != f2.Metadata.HasCmnt {
				return f1.Metadata.HasCmnt // files with comment first
			}
			return f1.Path < f2.Path
		case "name":
			fallthrough
		default:
			// Default to sorting by file name (or path if similar)
			// A true 'name' sort might only compare filepath.Base(Path).
			// Here we just use the rendered Path for stability, but we can do true name if needed.
			return f1.Path < f2.Path
		}
	})

	renderer := render.New(out, renderOpts)

	for _, fileInfo := range files {
		renderer.File(fileInfo)
	}

	renderer.Close()

	return 0
}
