import re

with open("internal/cli/commands.go", "r") as f:
    content = f.read()

run_list_sig_search = """func runList(runCfg *runOptions, mode string, recursive bool, filterExpr string, allXdg bool, allXattr bool, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, inspect bool, paths []string) error {"""
run_list_sig_replace = """func runList(runCfg *runOptions, mode string, recursive bool, filterExpr string, allXdg bool, allXattr bool, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, inspect bool, longListing bool, noGroup bool, noUser bool, showTitle bool, showAuthor bool, showCreator bool, showOrigin bool, showChecksum bool, showHidden bool, paths []string) error {"""
content = content.replace(run_list_sig_search, run_list_sig_replace)

render_opts_search = """	renderOpts := render.Options{
		JSONOutput:      jsonOutput,
		Inspect:         inspect || allXdg || allXattr,
		MaxTagsWidth:    maxTagsW,
		MaxCommentWidth: maxCmntW,
		NoWrap:          false,
		NoHeader:        noHeader,
	}"""
render_opts_replace = """	renderOpts := render.Options{
		JSONOutput:      jsonOutput,
		Inspect:         inspect || allXdg || allXattr,
		MaxTagsWidth:    maxTagsW,
		MaxCommentWidth: maxCmntW,
		NoWrap:          false,
		NoHeader:        noHeader,
		LongListing:     longListing,
		NoGroup:         noGroup,
		NoUser:          noUser,
		ShowTitle:       showTitle,
		ShowAuthor:      showAuthor,
		ShowCreator:     showCreator,
		ShowOrigin:      showOrigin,
		ShowChecksum:    showChecksum,
	}"""
content = content.replace(render_opts_search, render_opts_replace)

scan_opts_search = """	scanOpts := scanner.Options{
		Recursive: recursive,
		Filter:    finalFilter,
		XDGOnly:   xdgOnly,
		FS:        runCfg.fs,
	}"""
scan_opts_replace = """	scanOpts := scanner.Options{
		Recursive:  recursive,
		Filter:     finalFilter,
		XDGOnly:    xdgOnly,
		FS:         runCfg.fs,
		ShowHidden: showHidden,
	}"""
content = content.replace(scan_opts_search, scan_opts_replace)

lxa_search = """func Lxa(runCfg *runOptions, mode string, recursive bool, filterExpr string, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, paths ...string) error {
	return runList(runCfg, mode, recursive, filterExpr, false, false, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, false, paths)
}"""
lxa_replace = """func Lxa(runCfg *runOptions, mode string, recursive bool, filterExpr string, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, longListing bool, noGroup bool, noUser bool, showTitle bool, showAuthor bool, showCreator bool, showOrigin bool, showChecksum bool, showHidden bool, paths ...string) error {
	return runList(runCfg, mode, recursive, filterExpr, false, false, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, false, longListing, noGroup, noUser, showTitle, showAuthor, showCreator, showOrigin, showChecksum, showHidden, paths)
}"""
content = content.replace(lxa_search, lxa_replace)

inspect_search = """func Inspect(runCfg *runOptions, allXdg bool, allXattr bool, recursive bool, jsonOutput bool, maxTagsW int, maxCmntW int, sortField string, paths ...string) error {
	return runList(runCfg, "all", recursive, "", allXdg, allXattr, jsonOutput, false, maxTagsW, maxCmntW, sortField, true, paths)
}"""
inspect_replace = """func Inspect(runCfg *runOptions, allXdg bool, allXattr bool, recursive bool, jsonOutput bool, maxTagsW int, maxCmntW int, sortField string, paths ...string) error {
	return runList(runCfg, "all", recursive, "", allXdg, allXattr, jsonOutput, false, maxTagsW, maxCmntW, sortField, true, false, false, false, false, false, false, false, false, false, paths)
}"""
content = content.replace(inspect_search, inspect_replace)

with open("internal/cli/commands.go", "w") as f:
    f.write(content)
