package cli

import (
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"

	"github.com/lxa-project/lxa/internal/scanner"
	"github.com/lxa-project/lxa/internal/xattr"
)

//go:embed help.tmpl
var helpText string


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
	showHeader := false
	showAuthor := false
	showCreator := false
	showOrigin := false
	showChecksum := false
	showHidden := false
	singleColumn := false
	multiColumn := false

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
			case "header":
				showHeader = true
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
			case '1':
				singleColumn = true
			case 'C':
				multiColumn = true
			case 'X':
				allXdg = true
			case 'A':
				allXattr = true
			case 'm', 'f', 's', 'W', 'T':
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
				case 'W', 'T': // max-tags-width
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

	return Lxa(runCfg, mode, recursive, filterExpr, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, longListing, noGroup, noUser, showHeader, showAuthor, showCreator, showOrigin, showChecksum, showHidden, singleColumn, multiColumn, paths...)
}

func printHelp() {
	t, err := template.New("help").Parse(helpText)
	if err != nil {
		fmt.Fprintln(ErrOut, helpText)
		return
	}
	_ = t.Execute(ErrOut, nil)
}
