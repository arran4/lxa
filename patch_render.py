import re

with open("internal/render/render.go", "r") as f:
    content = f.read()

opts_struct_search = """type Options struct {
	JSONOutput      bool
	Inspect         bool
	MaxTagsWidth    int
	MaxCommentWidth int
	NoWrap          bool
	NoHeader        bool
}"""
opts_struct_replace = """type Options struct {
	JSONOutput      bool
	Inspect         bool
	MaxTagsWidth    int
	MaxCommentWidth int
	NoWrap          bool
	NoHeader        bool
	LongListing     bool
	NoGroup         bool
	NoUser          bool
	ShowTitle       bool
	ShowAuthor      bool
	ShowCreator     bool
	ShowOrigin      bool
	ShowChecksum    bool
}"""

content = content.replace(opts_struct_search, opts_struct_replace)


header_search = """	if !r.opts.NoHeader && !r.headerPrinted {
		_, _ = fmt.Fprintln(r.tw, "PERMISSIONS\\tNODE\\tOWNER\\tGROUP\\tSIZE\\tMODIFIED\\tFILENAME\\tTAGS\\tCOMMENTS")
		r.headerPrinted = true
	}"""
header_replace = """	if r.opts.ShowTitle && !r.opts.NoHeader && !r.headerPrinted {
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

		_, _ = fmt.Fprintln(r.tw, strings.Join(header, "\\t"))
		r.headerPrinted = true
	}"""

content = content.replace(header_search, header_replace)

list_render_search = """	_, _ = fmt.Fprintf(r.tw, "%s\\t%s\\t%s\\t%s\\t%s\\t%s\\t%s\\t%s\\t%s\\n", mode, node, owner, group, size, modTime, name, tagsStr, cmntStr)"""
list_render_replace = """	cols := []string{}
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

	_, _ = fmt.Fprintln(r.tw, strings.Join(cols, "\\t"))"""

content = content.replace(list_render_search, list_render_replace)


with open("internal/render/render.go", "w") as f:
    f.write(content)
