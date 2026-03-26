# lxa 1

## NAME

lxa - a Linux-first file listing and inspection tool focused on extended attributes and XDG metadata

## SYNOPSIS

**lxa** [*OPTIONS*] [*PATH*...]
**lxa inspect** [*OPTIONS*] [*PATH*...]

## DESCRIPTION

`lxa` is a file listing tool optimized specifically for answering: "Which files here have XDG metadata, and what is it?" It provides a compact, readable, terminal-width-aware output emphasizing XDG tags and comments alongside files. It can optionally display detailed long-listing information, file permissions, ownership, and other specialized user extended attributes.

## OPTIONS

**-m, --mode** *string*
: Filter mode: 'xdg', 'tags', 'comments', or 'all' (default "all").

**-R, --recursive**
: Traverse directories recursively.

**-f, --filter** *string*
: Apply filter expression. Supported logic: `and`, `or`, `not`, `()`. Example: `(tag:urgent or tag:projectX) and has:comment`.

**-j, --json**
: Output in JSON format.

**-H, --no-header**
: Do not print table headers.

**-l**
: Long listing format. Prints permissions, node, owner, group, size, and modified time.

**-o**
: Long listing without group information.

**-g**
: Long listing without user information.

**-a**
: Include hidden files (those starting with a dot `.`).

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

**-T, -W, --max-tags-width** *int*
: Maximum display width for tags (default 40).

**-C, --max-comment-width** *int*
: Maximum display width for comments (default 60). Note: When used alone, `-C` forces multi-column layout.

**-1**
: Single column layout.

**-C**
: Multi-column layout.

**-s, --sort** *string*
: Sort by: name, path, xdg, tags, comment (default "name").

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
