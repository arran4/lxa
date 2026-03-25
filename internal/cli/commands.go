package cli

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/lxa-project/lxa/internal/render"
	"github.com/lxa-project/lxa/internal/scanner"
	"github.com/lxa-project/lxa/internal/xattr"
)

var Out io.Writer = os.Stdout
var ErrOut io.Writer = os.Stderr
var FS scanner.FileSystem = nil
var XattrReader xattr.Reader = nil

func runList(mode string, recursive bool, filterExpr string, allXdg bool, allXattr bool, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, inspect bool, paths []string) error {
	finalFilter := filterExpr

	xdgOnly := false
	switch mode {
	case "xdg":
		xdgOnly = true
	case "tags":
		if finalFilter != "" {
			finalFilter = "(" + finalFilter + ") and has:tags"
		} else {
			finalFilter = "has:tags"
		}
	case "comments":
		if finalFilter != "" {
			finalFilter = "(" + finalFilter + ") and has:comment"
		} else {
			finalFilter = "has:comment"
		}
	}

	scanOpts := scanner.Options{
		Recursive: recursive,
		Filter:    finalFilter,
		XDGOnly:   xdgOnly,
		FS:        FS,
	}

	reader := XattrReader
	if reader == nil {
		reader = xattr.NewSyscallReader()
	}

	scan, err := scanner.New(reader, scanOpts)
	if err != nil {
		return fmt.Errorf("error initializing scanner: %w", err)
	}

	renderOpts := render.Options{
		JSONOutput:      jsonOutput,
		Inspect:         inspect || allXdg || allXattr,
		MaxTagsWidth:    maxTagsW,
		MaxCommentWidth: maxCmntW,
		NoWrap:          false,
		NoHeader:        noHeader,
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	ch := scan.Scan(paths)
	var files []scanner.FileInfo
	for fileInfo := range ch {
		files = append(files, fileInfo)
	}

	// Sorting
	sort.Slice(files, func(i, j int) bool {
		f1, f2 := files[i], files[j]

		switch sortField {
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
			return f1.Path < f2.Path
		}
	})

	renderer := render.New(Out, renderOpts)

	for _, fileInfo := range files {
		renderer.File(fileInfo)
	}

	renderer.Close()
	return nil
}

// Lxa is a subcommand `lxa` -- Lists files displaying extended attributes and XDG metadata
func Lxa(mode string, recursive bool, filterExpr string, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, paths ...string) error {
	return runList(mode, recursive, filterExpr, false, false, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, false, paths)
}

// Inspect is a subcommand `lxa inspect` -- Inspects file extended attributes in detail
func Inspect(allXdg bool, allXattr bool, recursive bool, jsonOutput bool, maxTagsW int, maxCmntW int, sortField string, paths ...string) error {
	return runList("all", recursive, "", allXdg, allXattr, jsonOutput, false, maxTagsW, maxCmntW, sortField, true, paths)
}
