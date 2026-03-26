1. **Define xattr Mutation Interface**:
   - Update `internal/xattr/xattr.go`: Rename `Reader` to `Store` (or add a new `Writer` interface) which includes `Set(path, name string, data []byte) error` and `Remove(path, name string) error`.
   - Implement `Set` and `Remove` in `SyscallReader` using `syscall.SYS_LSETXATTR` and `syscall.SYS_LREMOVEXATTR`.

2. **Enhance Metadata Parsing**:
   - Update `Metadata` struct in `internal/xattr/xattr.go` to hold parsed values for SELinux context, Samba DOS attributes, Capabilities presence, ACL presence, and rating (`user.xdg.rating` or `user.rating` or `user.baloo.rating` - I will use `user.baloo.rating` and `user.xdg.rating`, let's just use `user.baloo.rating` or `user.rating`? Since XDG tags/comments use `user.xdg.tags`, rating should probably be `user.xdg.rating`).
   - Parse `security.selinux`, `user.DOSATTRIB`, `security.capability`, `system.posix_acl_access`, `system.posix_acl_default` in `ReadMetadata`.

3. **Add CLI Flags for Display & Mutation**:
   - Update `internal/cli/main.go` and `internal/cli/help.tmpl`:
     - Display: `--selinux`, `--samba`, `--capabilities`, `--acl`.
     - Mutation: `--set-tags`, `--add-tags`, `--remove-tags`, `--clear-tags`, `--set-comment`, `--clear-comment`, `--set-rating`, `--clear-rating`.
   - Update `runOptions` to include a `xattrWriter` or generic `xattrStore`.

4. **Implement Mutation Logic**:
   - In `internal/cli/commands.go` (or a new file `mutate.go`), implement a mode that runs if mutation flags are present.
   - For each target path:
     - Read existing tags (if `add-tags` or `remove-tags` is used).
     - Apply tag mutations, deduplicate, and preserve order.
     - Set or clear `user.xdg.tags`, `user.xdg.comment`, `user.xdg.rating`.
     - Output success or failure for each file.
   - Wait, `lxa` might want to continue listing if display options are used? Or just exit. The standard behavior for such tools is usually that if you provide an action flag (like `chmod` or `chattr`), it applies it and exits. Let's make it exit after applying mutations.

5. **Update Renderers (List & Inspect)**:
   - `internal/render/render.go`:
     - Add `SELINUX`, `DOSATTRIB`, `CAPABILITIES`, `ACL` to long list header and formatting if their respective flags are set.
     - Update `renderInspect` to print these recognized families explicitly (e.g. `SELINUX Context: ...`, `Samba DOS Attributes: ...`).

6. **Testing**:
   - Update `internal/xattr/xattr_test.go` to test mocked Set/Remove and parsing of new attrs.
   - Add CLI parsing tests for mutations.
   - Test that mutation flags are mutually exclusive properly or handle sequences correctly.

7. **Pre-commit**:
   - Call `pre_commit_instructions` and follow them to "ensure proper testing, verification, review, and reflection are done".

8. **Submit**.
