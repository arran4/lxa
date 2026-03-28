# lxa 1

## NAME

lxa - a Linux-first file listing and inspection tool focused on extended attributes and XDG metadata

## SYNOPSIS

**lxa** [*OPTIONS*] [*PATH*...]
**lxa inspect** [*OPTIONS*] [*PATH*...]

## DESCRIPTION

`lxa` is a file listing tool optimized specifically for answering: "Which files here have XDG metadata, and what is it?" It provides a compact, readable, terminal-width-aware output emphasizing XDG tags and comments alongside files. It can optionally display detailed long-listing information, file permissions, ownership, and other specialized user extended attributes.

## OPTIONS

**--mode** *string*
: Filter mode: 'xdg', 'tags', 'comments', or 'all' (default "all").

**-m, --format** *string*
: Output format: 'commas'.

**-R, --recursive**
: Traverse directories recursively.

**--filter** *string*
: Apply filter expression. Supported logic: `and`, `or`, `not`, `()`. Example: `(tag:urgent or tag:projectX) and has:comment`.

**-f**
: Do not sort, enable -a.

**-j, --json**
: Output in JSON format.

**-H, --dereference-command-line**
: Follow symlinks on the command line.

**-l**
: Long listing format. Prints permissions, node, owner, group, size, and modified time.

**-o**
: Long listing without group information.

**-g**
: Long listing without user information.

**-a, --all**
: Include hidden files (those starting with a dot `.`).

**-A, --almost-all**
: Do not list implied . and ..

**-b, --escape**
: Print C-style escapes for nongraphic characters.

**-B, --ignore-backups**
: Do not list implied entries ending with ~.

**-c**
: Sort by ctime.

**-d, --directory**
: List directories themselves, not their contents.

**-D, --dired**
: Generate output designed for Emacs' dired mode.

**-F, --classify[=WHEN]**
: Append indicator.

**--file-type**
: Likewise, except do not append '*'.

**-G, --no-group**
: In a long listing, don't print group names.

**-h, --human-readable**
: Print sizes like 1K, 234M, 2G etc.

**--si**
: Likewise, but use powers of 1000 not 1024.

**-i, --inode**
: Print the index number of each file.

**-I, --ignore=PATTERN**
: Do not list implied entries matching shell PATTERN.

**-k, --kibibytes**
: Default to 1024-byte blocks for file system usage.

**-L, --dereference**
: When showing file info for symlink, show referenced file.

**-n, --numeric-uid-gid**
: List numeric user and group IDs.

**-N, --literal**
: Print entry names without quoting.

**-p, --indicator-style=slash**
: Append / indicator to directories.

**-q, --hide-control-chars**
: Print ? instead of nongraphic characters.

**-Q, --quote-name**
: Enclose entry names in double quotes.

**-r, --reverse**
: Reverse order while sorting.

**-s, --size**
: Print the allocated size of each file.

**-S**
: Sort by file size.

**--sort** *string*
: Sort by: name, size, time, version, extension, xdg, tags, comment (default "name").

**-t**
: Sort by time, newest first.

**-u**
: Sort by, and show, access time.

**-U**
: Do not sort directory entries.

**-v**
: Natural sort of (version) numbers.

**-w, --width=COLS**
: Set output width to COLS.

**-x**
: List entries by lines instead of by columns.

**-X**
: Sort alphabetically by entry extension.

**-Z, --context**
: Print any security context of each file.

**--zero**
: End each output line with NUL, not newline.

**--header**
: Show header row (title).

**--author**
: Show author (reads `user.author` xattr).

**--creator**
: Show creator (reads `user.creator` xattr).

**--origin**
: Show origin (reads `user.origin` xattr).

**--checksum**
: Show checksum (reads `user.checksum` xattr).

**--selinux**
: Show SELinux context (reads `security.selinux`).

**--samba**
: Show Samba DOS attributes (reads `user.DOSATTRIB`).

**--capabilities**
: Show file capabilities (reads `security.capability`).

**--acl**
: Show ACL presence (reads `system.posix_acl_access` and `system.posix_acl_default`).

**-W, --max-tags-width** *int*
: Maximum display width for tags (default 40).

**-1**
: Single column layout.

**-C**
: Multi-column layout. When used as `--max-comment-width`, sets comment width.

## INSPECT OPTIONS

**-X, --all-xdg**
: Show all XDG metadata attributes.

**-A, --all-xattr**
: Show all xattrs.

# MUTATION OPTIONS

**--set-tags=...**
: Set `user.xdg.tags` (comma-separated).

**--add-tags=...**
: Add tags without duplicates.

**--remove-tags=...**
: Remove specified tags.

**--clear-tags**
: Remove `user.xdg.tags` entirely.

**--set-comment=...**
: Set `user.xdg.comment`.

**--clear-comment**
: Remove `user.xdg.comment` entirely.

**--set-rating=...**
: Set `user.xdg.rating`.

**--clear-rating**
: Remove `user.xdg.rating` entirely.

## EXAMPLES

List files with XDG metadata prominently displayed:
```
$ lxa
```

List files with long listing format and headers:
```
$ lxa -l --header
```

Show only files with XDG metadata, tags, or comments specifically using the `-m` flag:
```
$ lxa -m tags
```

List files that have not been seen:
```
$ lxa --filter 'not tag:Seen'
```

Inspect detailed xattrs for a specific file:
```
$ lxa inspect myfile.txt
```

## BUGS

Report bugs at https://github.com/lxa-project/lxa/issues
