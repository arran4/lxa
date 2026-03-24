# lxa

`lxa` is a Linux-first file listing and inspection tool focused on extended attributes, especially XDG metadata (`user.xdg.tags`, `user.xdg.comment`, etc).

## Why `lxa`?

You already have `ls`, but `ls` doesn't expose XDG metadata well. `lxa` isn't designed to be a better general `ls`. It's optimized specifically for answering:
*"Which files here have XDG metadata, and what is it?"*

It provides a compact, readable, terminal-width-aware output emphasizing XDG tags and comments.

## Features

- **XDG Focus**: Easily see XDG tags and comments alongside files.
- **Filtering**: Powerful expressions to find files (e.g. `has:tags and not has:comment`, `tag:urgent`).
- **Width-Aware**: Intelligently truncates tags and comments to fit your terminal.
- **Inspect Mode**: Detailed view of all `user.xdg.*` and other extended attributes.
- **JSON Output**: Structured, parseable JSON for scripting.
- **No Dependencies**: Pure Go, native Linux syscalls, no shelling out to `getfattr`.

## Installation

### Build from source
```bash
go install github.com/lxa-project/lxa/cmd/lxa@latest
```

## Usage

### Basic List
List files with XDG metadata prominently displayed:
```bash
$ lxa
file1.txt
file2.txt [tags: projectX] [comment: needs review]
```

### Modes
Show only files with XDG metadata, tags, or comments specifically using the `-m` (mode) flag:
```bash
$ lxa -m tags
file2.txt [tags: projectX] [comment: needs review]
```

### Recursive traversal
```bash
lxa -R
```

### Filtering Expressions
Use expressions to find specific files. Supported logic: `and`, `or`, `not`, `()`.
```bash
$ lxa --filter "(tag:urgent or tag:projectX) and has:comment"
file2.txt [tags: projectX] [comment: needs review]
```

### Inspect Mode
Inspect detailed xattrs for a specific file:
```bash
$ lxa inspect myfile.txt
Path: myfile.txt
  Type: -rw-r--r--
  Size: 2048
  XDG Metadata:
    tags: projectX, review
    comment: urgent task

  Other xattrs:
    user.custom: info
```

### JSON Output
```bash
$ lxa --json myfile.txt
[
  {
    "metadata": {
      "All": {
        "user.xdg.tags": "cHJvamVjdFgsIHJldmlldw=="
      },
      "Comment": "",
      "HasCmnt": false,
      "HasTags": true,
      "HasXDG": true,
      "Tags": [
        "projectX",
        "review"
      ],
      "XDG": {
        "user.xdg.tags": "cHJvamVjdFgsIHJldmlldw=="
      }
    },
    "path": "myfile.txt",
    "size": 2048,
    "type": "-rw-r--r--"
  }
]
```

## Supported Filters

Filters allow composing powerful searches:

- `xdg` - Has any `user.xdg.*` metadata
- `has:tags` - Has `user.xdg.tags`
- `has:comment` - Has `user.xdg.comment`
- `tag:foo` - Has tag containing "foo"
- `xdg:key` - Has specific XDG key `user.xdg.key`
- `xattr:key` - Has arbitrary xattr `key`

Example:
```bash
lxa --filter '(tag:urgent or tag:projectX) and has:comment'
```
