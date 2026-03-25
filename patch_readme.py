import re

with open("README.md", "r") as f:
    content = f.read()

usage_search = """$ lxa -l
PERMISSIONS  NODE  OWNER  GROUP  SIZE  MODIFIED      FILENAME                    TAGS  COMMENTS"""
usage_replace = """$ lxa -l -T
PERMISSIONS  NODE  OWNER  GROUP  SIZE  MODIFIED      FILENAME                    TAGS  COMMENTS"""
content = content.replace(usage_search, usage_replace)

features_search = """- **No Dependencies**: Pure Go, native Linux syscalls, no shelling out to `getfattr`."""
features_replace = """- **No Dependencies**: Pure Go, native Linux syscalls, no shelling out to `getfattr`.
- **Customizable Output**: Long listing formats (`-l`, `-o`, `-g`), toggle headers (`-T`), and specialized xattr columns (`--author`, `--checksum`, etc.)."""
content = content.replace(features_search, features_replace)

with open("README.md", "w") as f:
    f.write(content)
