package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/lxa-project/lxa/internal/render"
	"github.com/lxa-project/lxa/internal/scanner"
	"github.com/lxa-project/lxa/internal/xattr"
)

var Out io.Writer = os.Stdout
var ErrOut io.Writer = os.Stderr

func runList(runCfg *runOptions, mode string, recursive bool, filterExpr string, allXdg bool, allXattr bool, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, inspect bool, longListing bool, noGroup bool, noUser bool, showHeader bool, showAuthor bool, showCreator bool, showOrigin bool, showChecksum bool, showSELinux bool, showSamba bool, showCapabilities bool, showACL bool, showHidden bool, singleColumn bool, multiColumn bool, paths []string) error {
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
		Recursive:  recursive,
		Filter:     finalFilter,
		XDGOnly:    xdgOnly,
		FS:         runCfg.fs,
		ShowHidden: showHidden,
	}

	var reader xattr.Reader
	if runCfg.xattrStore != nil {
		reader = runCfg.xattrStore
	} else {
		reader = xattr.NewSyscallReader().(xattr.Store)
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
		LongListing:     longListing,
		NoGroup:         noGroup,
		NoUser:          noUser,
		ShowHeader:      showHeader,
		ShowAuthor:      showAuthor,
		ShowCreator:     showCreator,
		ShowOrigin:      showOrigin,
		ShowChecksum:     showChecksum,
		ShowSELinux:      showSELinux,
		ShowSamba:        showSamba,
		ShowCapabilities: showCapabilities,
		ShowACL:          showACL,
		SingleColumn:     singleColumn,
		MultiColumn:      multiColumn,
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
func Lxa(runCfg *runOptions, mode string, recursive bool, filterExpr string, jsonOutput bool, noHeader bool, maxTagsW int, maxCmntW int, sortField string, longListing bool, noGroup bool, noUser bool, showHeader bool, showAuthor bool, showCreator bool, showOrigin bool, showChecksum bool, showSELinux bool, showSamba bool, showCapabilities bool, showACL bool, showHidden bool, singleColumn bool, multiColumn bool, paths ...string) error {
	return runList(runCfg, mode, recursive, filterExpr, false, false, jsonOutput, noHeader, maxTagsW, maxCmntW, sortField, false, longListing, noGroup, noUser, showHeader, showAuthor, showCreator, showOrigin, showChecksum, showSELinux, showSamba, showCapabilities, showACL, showHidden, singleColumn, multiColumn, paths)
}

// Inspect is a subcommand `lxa inspect` -- Inspects file extended attributes in detail
func Inspect(runCfg *runOptions, allXdg bool, allXattr bool, recursive bool, jsonOutput bool, maxTagsW int, maxCmntW int, sortField string, paths ...string) error {
	return runList(runCfg, "all", recursive, "", allXdg, allXattr, jsonOutput, false, maxTagsW, maxCmntW, sortField, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, paths)
}

func runMutate(runCfg *runOptions, paths []string, setTags, addTags, removeTags *string, clearTags bool, setComment *string, clearComment bool, setRating *string, clearRating bool) error {
	// Mutually exclusive flags validation
	if clearTags && (setTags != nil || addTags != nil || removeTags != nil) {
		return fmt.Errorf("cannot use --clear-tags with other tag mutation flags")
	}
	if setTags != nil && (addTags != nil || removeTags != nil) {
		return fmt.Errorf("cannot use --set-tags with --add-tags or --remove-tags")
	}
	if clearComment && setComment != nil {
		return fmt.Errorf("cannot use --clear-comment with --set-comment")
	}
	if clearRating && setRating != nil {
		return fmt.Errorf("cannot use --clear-rating with --set-rating")
	}

	// Validate rating value
	if setRating != nil {
		_, err := strconv.Atoi(*setRating)
		if err != nil {
			return fmt.Errorf("invalid rating value %q: must be an integer", *setRating)
		}
	}

	var store xattr.Store
	if runCfg.xattrStore != nil {
		store = runCfg.xattrStore
	} else {
		store = xattr.NewSyscallReader().(xattr.Store)
	}

	for _, path := range paths {
		if clearTags {
			err := store.Remove(path, "user.xdg.tags")
			if err != nil {
				return fmt.Errorf("failed to clear tags on %s: %w", path, err)
			}
		} else if setTags != nil {
			err := store.Set(path, "user.xdg.tags", []byte(*setTags+"\x00"))
			if err != nil {
				return fmt.Errorf("failed to set tags on %s: %w", path, err)
			}
		} else if addTags != nil || removeTags != nil {
			md, err := xattr.ReadMetadata(store, path)
			if err != nil {
				return fmt.Errorf("failed to read metadata on %s: %w", path, err)
			}
			tags := md.Tags
			if removeTags != nil {
					toRemove := strings.Split(*removeTags, ",")
					var newTags []string
					for _, t := range tags {
						remove := false
						for _, r := range toRemove {
							if strings.TrimSpace(r) == t {
								remove = true
								break
							}
						}
						if !remove {
							newTags = append(newTags, t)
						}
					}
					tags = newTags
				}
				if addTags != nil {
					toAdd := strings.Split(*addTags, ",")
					for _, a := range toAdd {
						a = strings.TrimSpace(a)
						exists := false
						for _, t := range tags {
							if t == a {
								exists = true
								break
							}
						}
						if !exists && a != "" {
							tags = append(tags, a)
						}
					}
				}
			if len(tags) > 0 {
				err = store.Set(path, "user.xdg.tags", []byte(strings.Join(tags, ",")+"\x00"))
				if err != nil {
					return fmt.Errorf("failed to update tags on %s: %w", path, err)
				}
			} else if md.HasTags {
				err = store.Remove(path, "user.xdg.tags")
				if err != nil {
					return fmt.Errorf("failed to remove tags on %s: %w", path, err)
				}
			}
		}

		if clearComment {
			err := store.Remove(path, "user.xdg.comment")
			if err != nil {
				return fmt.Errorf("failed to clear comment on %s: %w", path, err)
			}
		} else if setComment != nil {
			err := store.Set(path, "user.xdg.comment", []byte(*setComment+"\x00"))
			if err != nil {
				return fmt.Errorf("failed to set comment on %s: %w", path, err)
			}
		}

		if clearRating {
			err := store.Remove(path, "user.xdg.rating")
			if err != nil {
				return fmt.Errorf("failed to clear rating on %s: %w", path, err)
			}
		} else if setRating != nil {
			err := store.Set(path, "user.xdg.rating", []byte(*setRating+"\x00"))
			if err != nil {
				return fmt.Errorf("failed to set rating on %s: %w", path, err)
			}
		}
	}

	return nil
}
