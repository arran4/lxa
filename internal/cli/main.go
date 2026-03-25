package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/lxa-project/lxa/internal/scanner"
	"github.com/lxa-project/lxa/internal/xattr"
)

type runOptions struct {
	fs          scanner.FileSystem
	xattrReader xattr.Reader
}

type RunOption func(*runOptions)

func WithFS(fs scanner.FileSystem) RunOption {
	return func(o *runOptions) {
		o.fs = fs
	}
}

func WithXattrReader(r xattr.Reader) RunOption {
	return func(o *runOptions) {
		o.xattrReader = r
	}
}

// Run parses flags manually supporting combined short flags (e.g. -Rj)
// It handles "inspect" as a proper subcommand.
func Run(args []string, out io.Writer, errOut io.Writer, opts ...RunOption) error {
	runCfg := &runOptions{}
	for _, opt := range opts {
		opt(runCfg)
	}

	Out = out
	ErrOut = errOut

	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		printHelp()
		return nil
	}

	mode := "all"
	recursive := false
	filterExpr := ""
	jsonOutput := false
	maxTagsW := 40
	maxCmntW := 60
	sortField := "name"
	noHeader := false

	longListing := false
	noGroup := false
	noUser := false
	showTitle := false
	showAuthor := false
	showCreator := false
	showOrigin := false
	showChecksum := false
	showHidden := false

	paths := []string{}

	allXdg := false
	allXattr := false

	inspectMode := false

	if len(args) > 0 && args[0] == "inspect" {
		inspectMode = true
		args = args[1:]
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			paths = append(paths, args[i+1:]...)
			break
		}

		if !strings.HasPrefix(arg, "-") {
			paths = append(paths, arg)
			continue
		}

		if strings.HasPrefix(arg, "--") {
			name := strings.TrimPrefix(arg, "--")
			val := ""
			hasVal := false

			if idx := strings.Index(name, "="); idx >= 0 {
				val = name[idx+1:]
				name = name[:idx]
				hasVal = true
			}

			switch name {
			case "mode":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				mode = val
			case "filter":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				filterExpr = val
			case "max-tags-width":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				maxTagsW, _ = strconv.Atoi(val)
			case "title":
				showTitle = true
			case "max-comment-width":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				maxCmntW, _ = strconv.Atoi(val)
			case "sort":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				sortField = val
			case "recursive":
				recursive = true
			case "json":
				jsonOutput = true
			case "no-header":
				noHeader = true
			case "author":
				showAuthor = true
			case "creator":
				showCreator = true
			case "origin":
				showOrigin = true
			case "checksum":
				showChecksum = true
			case "all-xdg":
				allXdg = true
			case "all-xattr":
				allXattr = true
			default:
				return fmt.Errorf("unknown flag: %s", arg)
			}
			continue
		}

		// Short flags (can be combined like -Rj)
		chars := arg[1:]
		for j, c := range chars {
			switch c {
			case 'R':
				recursive = true
			case 'j':
				jsonOutput = true
			case 'H':
				noHeader = true
			case 'l':
				longListing = true
			case 'o':
				longListing = true
				noGroup = true
			case 'g':
				longListing = true
				noUser = true
			case 'a':
				showHidden = true
			case 'X':
				allXdg = true
			case 'A':
				allXattr = true
			case 'T':
				showTitle = true
			case 'm', 'f', 'C', 's', 'W':
				val := ""
				if j+1 < len(chars) {
					// rest of string is the value
					val = string(chars[j+1:])
				} else if i+1 < len(args) {
					// next arg is the value
					i++
					val = args[i]
				}

				switch c {
				case 'm':
					mode = val
				case 'f':
					filterExpr = val
				case 'W': // max-tags-width
					maxTagsW, _ = strconv.Atoi(val)
				case 'C':
					maxCmntW, _ = strconv.Atoi(val)
				case 's':
					sortField = val
				}
				goto nextArg
			default:
				return fmt.Errorf("unknown short flag: -%c", c)
			}
		}
	nextArg:
	}

	if inspectMode {
		return Inspect(runCfg, allXdg, allXattr, recursive, jsonOutput, maxTagsW, maxCmntW, sortField, paths...)
	}

	return Lxa(runCfg, mode, recursive, filterExpr, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, longListing, noGroup, noUser, showTitle, showAuthor, showCreator, showOrigin, showChecksum, showHidden, paths...)
}

func printHelp() {
	fmt.Fprintln(ErrOut, "Usage: lxa [OPTIONS] [PATH...]")
	fmt.Fprintln(ErrOut, "       lxa inspect [PATH...]")
	fmt.Fprintln(ErrOut, "\nlxa is a Linux-first file listing tool focused on extended attributes and XDG metadata.")
	fmt.Fprintln(ErrOut, "\nOptions:")
	fmt.Fprintln(ErrOut, "  -m, --mode string              Filter mode: 'xdg', 'tags', 'comments', or 'all' (default \"all\")")
	fmt.Fprintln(ErrOut, "  -R, --recursive                Traverse directories recursively")
	fmt.Fprintln(ErrOut, "  -f, --filter string            Apply filter expression")
	fmt.Fprintln(ErrOut, "  -j, --json                     Output in JSON format")
	fmt.Fprintln(ErrOut, "  -H, --no-header                Do not print table headers")
	fmt.Fprintln(ErrOut, "  -l                             Long listing format")
	fmt.Fprintln(ErrOut, "  -o                             Long listing without group information")
	fmt.Fprintln(ErrOut, "  -g                             Long listing without user information")
	fmt.Fprintln(ErrOut, "  -a                             Include hidden files")
	fmt.Fprintln(ErrOut, "  -T, --title                    Show title (header row)")
	fmt.Fprintln(ErrOut, "      --author                   Show author (user.author)")
	fmt.Fprintln(ErrOut, "      --creator                  Show creator (user.creator)")
	fmt.Fprintln(ErrOut, "      --origin                   Show origin (user.origin)")
	fmt.Fprintln(ErrOut, "      --checksum                 Show checksum (user.checksum)")
	fmt.Fprintln(ErrOut, "  -W, --max-tags-width int       Maximum display width for tags (default 40)")
	fmt.Fprintln(ErrOut, "  -C, --max-comment-width int    Maximum display width for comments (default 60)")
	fmt.Fprintln(ErrOut, "  -s, --sort string              Sort by: name, path, xdg, tags, comment (default \"name\")")
	fmt.Fprintln(ErrOut, "\nInspect Options:")
	fmt.Fprintln(ErrOut, "  -X, --all-xdg                  Show all XDG metadata attributes")
	fmt.Fprintln(ErrOut, "  -A, --all-xattr                Show all xattrs")
}
