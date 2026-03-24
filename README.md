# lxa

`lxa` is a Linux-first file listing and inspection tool focused on extended attributes, especially XDG metadata (`user.xdg.tags`, `user.xdg.comment`, etc).

## Why `lxa`?

You already have `ls`, but `ls` doesn't expose XDG metadata well. `lxa` isn't designed to be a better general `ls`. It's optimized specifically for answering:
*"Which files here have XDG metadata, and what is it?"*

It provides a compact, readable, terminal-width-aware output emphasizing XDG tags and comments.

## Features

- **XDG Focus**: Easily see XDG tags and comments alongside files.
- **Filtering**: Powerful expressions to find files (e.g. `has:tags && !has:comment`, `tag:urgent`).
- **Width-Aware**: Intelligently truncates tags and comments to fit your terminal.
- **Inspect Mode**: Detailed view of all `user.xdg.*` and other extended attributes.
- **JSON Output**: Structured, parseable JSON for scripting.
- **No Dependencies**: Pure Go, native Linux syscalls, no shelling out to `getfattr`.

## Installation

Download a pre-built binary from the [Releases](https://github.com/lxa-project/lxa/releases) page.

Alternatively, compile from source:

```bash
go install github.com/lxa-project/lxa/cmd/lxa@latest
```

## Usage

List files with XDG metadata prominently displayed:
```bash
lxa
```

Show only files with XDG metadata:
```bash
lxa --xdg-only
```

Use expressions to filter files:
```bash
lxa --filter "tag:urgent || tag:projectX"
```

Inspect detailed xattrs for a specific file:
```bash
lxa inspect myfile.txt
```

JSON output:
```bash
lxa --json
```

## Supported Filters

Filters allow composing powerful searches:

- `xdg` - Has any `user.xdg.*` metadata
- `has:tags` - Has `user.xdg.tags`
- `has:comment` - Has `user.xdg.comment`
- `tag:foo` - Has tag containing "foo"
- `xdg:key` - Has specific XDG key `user.xdg.key`
- `xattr:key` - Has arbitrary xattr `key`

Logical operators supported: `&&`, `||`, `!`, and grouping `()`.

Example:
```bash
lxa --filter '(tag:urgent || tag:projectX) && has:comment'
```
