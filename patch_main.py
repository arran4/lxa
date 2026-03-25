import re

with open("internal/cli/main.go", "r") as f:
    content = f.read()

flags_search = """	mode := "all"
	recursive := false
	filterExpr := ""
	jsonOutput := false
	maxTagsW := 40
	maxCmntW := 60
	sortField := "name"
	noHeader := false
	paths := []string{}

	allXdg := false
	allXattr := false"""
flags_replace = """	mode := "all"
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
	allXattr := false"""
content = content.replace(flags_search, flags_replace)

long_args_search = """			case "no-header":
				noHeader = true"""
long_args_replace = """			case "no-header":
				noHeader = true
			case "author":
				showAuthor = true
			case "creator":
				showCreator = true
			case "origin":
				showOrigin = true
			case "checksum":
				showChecksum = true"""
content = content.replace(long_args_search, long_args_replace)

short_args_search = """			case 'H':
				noHeader = true"""
short_args_replace = """			case 'H':
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
				showHidden = true"""
content = content.replace(short_args_search, short_args_replace)

t_arg_search = """			case 'm', 'f', 'T', 'C', 's':
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
				case 'T':
					maxTagsW, _ = strconv.Atoi(val)"""
t_arg_replace = """			case 'T':
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
					maxTagsW, _ = strconv.Atoi(val)"""
content = content.replace(t_arg_search, t_arg_replace)

t_long_arg_search = """			case "max-tags-width":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				maxTagsW, _ = strconv.Atoi(val)"""
t_long_arg_replace = """			case "max-tags-width":
				if !hasVal && i+1 < len(args) {
					i++
					val = args[i]
				}
				maxTagsW, _ = strconv.Atoi(val)
			case "title":
				showTitle = true"""
content = content.replace(t_long_arg_search, t_long_arg_replace)

lxa_call_search = """	return Lxa(runCfg, mode, recursive, filterExpr, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, paths...)"""
lxa_call_replace = """	return Lxa(runCfg, mode, recursive, filterExpr, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, longListing, noGroup, noUser, showTitle, showAuthor, showCreator, showOrigin, showChecksum, showHidden, paths...)"""
content = content.replace(lxa_call_search, lxa_call_replace)

help_search = """	fmt.Fprintln(ErrOut, "  -H, --no-header                Do not print table headers")
	fmt.Fprintln(ErrOut, "  -T, --max-tags-width int       Maximum display width for tags (default 40)")"""
help_replace = """	fmt.Fprintln(ErrOut, "  -H, --no-header                Do not print table headers")
	fmt.Fprintln(ErrOut, "  -l                             Long listing format")
	fmt.Fprintln(ErrOut, "  -o                             Long listing without group information")
	fmt.Fprintln(ErrOut, "  -g                             Long listing without user information")
	fmt.Fprintln(ErrOut, "  -a                             Include hidden files")
	fmt.Fprintln(ErrOut, "  -T, --title                    Show title (header row)")
	fmt.Fprintln(ErrOut, "      --author                   Show author (user.author)")
	fmt.Fprintln(ErrOut, "      --creator                  Show creator (user.creator)")
	fmt.Fprintln(ErrOut, "      --origin                   Show origin (user.origin)")
	fmt.Fprintln(ErrOut, "      --checksum                 Show checksum (user.checksum)")
	fmt.Fprintln(ErrOut, "  -W, --max-tags-width int       Maximum display width for tags (default 40)")"""
content = content.replace(help_search, help_replace)

with open("internal/cli/main.go", "w") as f:
    f.write(content)
