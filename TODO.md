# TODO

## Planning & Design
- [x] Read existing codebase (cli, render, xattr).
- [x] Write design note.

## 1. xattr writing
- [ ] Extend `xattr.Reader` into a `xattr.ReadWriter` or add `Writer` interface.
- [ ] Implement `Set` and `Remove` in `SyscallReader` using `lsetxattr` and `lremovexattr`.

## 2. Metadata struct updates
- [ ] Add fields for SELinux context (`security.selinux`), Samba DOS summary (`user.DOSATTRIB`), Capability presence (`security.capability`), ACL presence (`system.posix_acl_access`/`default`), Rating (`user.xdg.rating` or similar, perhaps check XDG rating convention? Baloo uses `user.baloo.rating` but this is an XDG tool. I'll use `user.baloo.rating` as well as `user.xdg.rating`?).
- [ ] Actually, XDG rating might just be `user.xdg.rating`? I will use `user.xdg.rating` or ask in design note. Let's use `user.xdg.rating` for now.
- [ ] Implement parsing logic in `ReadMetadata` in `xattr.go`.

## 3. CLI updates
- [ ] Add display flags: `--selinux`, `--samba`, `--capabilities`, `--acl`.
- [ ] Add mutation flags: `--set-tags`, `--add-tags`, `--remove-tags`, `--clear-tags`, `--set-comment`, `--clear-comment`, `--set-rating`, `--clear-rating`.
- [ ] Add mutation mode handling (execute on paths, exit).

## 4. Render updates
- [ ] Update `renderList` to conditionally include SELinux, Samba, Cap, ACL columns.
- [ ] Update `renderInspect` to format recognized sections (SELinux, Samba, Capabilities, ACL).

## 5. Mutation commands execution
- [ ] `cmdMutate`: handle combination of operations safely.
- [ ] Test mutation commands on mocked fs/xattr.

## 6. Pre-commit & Tests
- [ ] Add tests for new CLI flags.
- [ ] Add tests for rendering output.
- [ ] Add tests for write/mutation commands.
- [ ] Complete pre-commit.
