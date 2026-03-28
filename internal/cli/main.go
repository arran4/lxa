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

var Version = "dev"

type runOptions struct {
	fs          scanner.FileSystem
	xattrStore  xattr.Store
}

type RunOption func(*runOptions)

func WithFS(fs scanner.FileSystem) RunOption {
	return func(o *runOptions) {
		o.fs = fs
	}
}

func WithXattrStore(s xattr.Store) RunOption {
	return func(o *runOptions) {
		o.xattrStore = s
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

	if len(args) > 0 && (args[0] == "-v" || args[0] == "--version" || args[0] == "version") {
		fmt.Fprintf(Out, "lxa version %s\n", Version)
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
	showSELinux := false
	showSamba := false
	showCapabilities := false
	showACL := false
	showHidden := false
	singleColumn := false
	multiColumn := false

	almostAll := false
	escape := false
	blockSize := ""
	ignoreBackups := false
	directory := false
	dired := false
	classify := ""
	fileType := false
	format := ""
	fullTime := false
	groupDirsFirst := false
	humanReadable := false
	si := false
	dereferenceCmdLine := false
	dereferenceCmdLineDir := false
	hidePattern := ""
	hyperlink := ""
	indicatorStyle := ""
	inode := false
	ignorePattern := ""
	kibibytes := false
	dereference := false
	numericUidGid := false
	literal := false
	indicatorSlash := false
	hideControlChars := false
	showControlChars := false
	quoteName := false
	quotingStyle := ""
	reverse := false
	allocSize := false
	timeWord := ""
	timeStyle := ""
	tabsize := 8
	widthCols := 0
	lines := false
	context := false
	zero := false

	var setTags, addTags, removeTags, setComment, setRating *string
	clearTags := false
	clearComment := false
	clearRating := false

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
				if !hasVal && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
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
			case "selinux":
				showSELinux = true
			case "samba":
				showSamba = true
			case "capabilities":
				showCapabilities = true
			case "acl":
				showACL = true
			case "set-tags":
				v := val
				setTags = &v
			case "add-tags":
				v := val
				addTags = &v
			case "remove-tags":
				v := val
				removeTags = &v
			case "clear-tags":
				clearTags = true
			case "set-comment":
				v := val
				setComment = &v
			case "clear-comment":
				clearComment = true
			case "set-rating":
				v := val
				setRating = &v
			case "clear-rating":
				clearRating = true
			case "all-xdg":
				allXdg = true
			case "all-xattr":
				allXattr = true
			case "almost-all":
				almostAll = true
			case "escape":
				escape = true
			case "block-size":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				blockSize = val
			case "ignore-backups":
				ignoreBackups = true
			case "directory":
				directory = true
			case "dired":
				dired = true
			case "classify":
				if !hasVal && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					val = args[i]
				}
				if val == "" {
					val = "always"
				}
				classify = val
			case "file-type":
				fileType = true
			case "format":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				format = val
			case "full-time":
				fullTime = true
			case "group-directories-first":
				groupDirsFirst = true
			case "no-group":
				noGroup = true
			case "human-readable":
				humanReadable = true
			case "si":
				si = true
			case "dereference-command-line":
				dereferenceCmdLine = true
			case "dereference-command-line-symlink-to-dir":
				dereferenceCmdLineDir = true
			case "hide":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				hidePattern = val
			case "hyperlink":
				if !hasVal && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					val = args[i]
				}
				if val == "" {
					val = "always"
				}
				hyperlink = val
			case "indicator-style":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				indicatorStyle = val
			case "inode":
				inode = true
			case "ignore":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				ignorePattern = val
			case "kibibytes":
				kibibytes = true
			case "dereference":
				dereference = true
			case "numeric-uid-gid":
				numericUidGid = true
			case "literal":
				literal = true
			case "hide-control-chars":
				hideControlChars = true
			case "show-control-chars":
				showControlChars = true
			case "quote-name":
				quoteName = true
			case "quoting-style":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				quotingStyle = val
			case "reverse":
				reverse = true
			case "size":
				allocSize = true
			case "time":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				timeWord = val
			case "time-style":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				timeStyle = val
			case "tabsize":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				tabsize, _ = strconv.Atoi(val)
			case "width":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				widthCols, _ = strconv.Atoi(val)
			case "context":
				context = true
			case "zero":
				zero = true
			case "all": // mapping for -a
				showHidden = true
			default:
				return fmt.Errorf("unknown flag: %s\nRun 'lxa --help' for usage.", arg)
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
				dereferenceCmdLine = true // Note: Originally mapped to no-header, changing meaning based on GNU ls, so we need to fix it. Wait, GNU ls uses -H for dereference-command-line, so we should map -H to dereferenceCmdLine. But lxa previously mapped -H to no-header. Let's map -H to dereferenceCmdLine to comply with GNU ls and user request. But how about --no-header?
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
			case 'A':
				almostAll = true
			case 'b':
				escape = true
			case 'B':
				ignoreBackups = true
			case 'c':
				timeWord = "ctime"
				// Note: technically -c affects sorting, we can handle it later in runList
			case 'd':
				directory = true
			case 'D':
				dired = true
			case 'f':
				showHidden = true
				sortField = "none"
			case 'F':
				classify = "always"
			case 'h':
				humanReadable = true
			case 'i':
				inode = true
			case 'k':
				kibibytes = true
			case 'L':
				dereference = true
			case 'm':
				format = "commas"
			case 'n':
				longListing = true
				numericUidGid = true
			case 'N':
				literal = true
			case 'p':
				indicatorSlash = true
			case 'q':
				hideControlChars = true
			case 'Q':
				quoteName = true
			case 'r':
				reverse = true
			case 's':
				allocSize = true
			case 'S':
				sortField = "size"
			case 't':
				sortField = "time"
			case 'u':
				timeWord = "atime"
			case 'U':
				sortField = "none"
			case 'v':
				sortField = "version"
			case 'x':
				format = "horizontal"
			case 'X':
				sortField = "extension"
			case 'Z':
				context = true
			case 'C':
				multiColumn = true
			case 'W': // keeping Lxa original max-tags-width
				val := ""
				if j+1 < len(chars) {
					val = string(chars[j+1:])
				} else if i+1 < len(args) {
					i++
					val = args[i]
				}
				maxTagsW, _ = strconv.Atoi(val)
				goto nextArg
			case 'I', 'w', 'T': // -m and -s and -f are repurposed, so we'll need to use long flags for mode/filter now, or redefine them. GNU ls uses -m for commas, -s for size, -f for not sort. Wait, original lxa used -m for mode, -f for filter, -s for sort, -W/-T for max-tags-width, -C for max-comment-width. The prompt said: "do not ignore entries starting with ." ... "Also error on invalid flags. Create a todo list and implement each one checking off the todo list then clean up." The user requested GNU ls flags compatibility. If there's a conflict between lxa specific flags and GNU ls flags, I should prioritize GNU ls flags because the prompt states "All of the flags visible need to be implemented".
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
				case 'I':
					ignorePattern = val
				case 'w':
					widthCols, _ = strconv.Atoi(val)
				case 'T':
					tabsize, _ = strconv.Atoi(val)
				}
				goto nextArg
			default:
				return fmt.Errorf("unknown short flag: -%c\nRun 'lxa --help' for usage.", c)
			}
		}
	nextArg:
	}

	hasMutation := setTags != nil || addTags != nil || removeTags != nil || clearTags || setComment != nil || clearComment || setRating != nil || clearRating
	if hasMutation {
		return runMutate(runCfg, paths, setTags, addTags, removeTags, clearTags, setComment, clearComment, setRating, clearRating)
	}

	if inspectMode {
		return Inspect(runCfg, allXdg, allXattr, recursive, jsonOutput, maxTagsW, maxCmntW, sortField, paths...)
	}

	return Lxa(runCfg, mode, recursive, filterExpr, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, longListing, noGroup, noUser, showHeader, showAuthor, showCreator, showOrigin, showChecksum, showSELinux, showSamba, showCapabilities, showACL, showHidden, singleColumn, multiColumn, almostAll, escape, blockSize, ignoreBackups, directory, dired, classify, fileType, format, fullTime, groupDirsFirst, humanReadable, si, dereferenceCmdLine, dereferenceCmdLineDir, hidePattern, hyperlink, indicatorStyle, inode, ignorePattern, kibibytes, dereference, numericUidGid, literal, indicatorSlash, hideControlChars, showControlChars, quoteName, quotingStyle, reverse, allocSize, timeWord, timeStyle, tabsize, widthCols, lines, context, zero, paths...)
}

func printHelp() {
	t, err := template.New("help").Parse(helpText)
	if err != nil {
		fmt.Fprintln(ErrOut, helpText)
		return
	}
	_ = t.Execute(ErrOut, nil)
}
