# Design Note for lxa Attribute Expansion

## 1. Public CLI flags proposed

### Display Options
- `--selinux`: Shows SELinux context in listing.
- `--samba`: Shows Samba DOS attributes summary in listing.
- `--capabilities`: Shows file capabilities in listing.
- `--acl`: Shows ACL presence in listing.
- Existing flags (`--author`, `--creator`, `--origin`, `--checksum`) are retained.

### Mutation Options (Action flags)
- `--set-tags=tag1,tag2`: Overwrites tags with the specified list.
- `--add-tags=tag1,tag2`: Adds tags without duplicates.
- `--remove-tags=tag1,tag2`: Removes specified tags.
- `--clear-tags`: Removes the `user.xdg.tags` attribute.
- `--set-comment=TEXT`: Sets `user.xdg.comment`.
- `--clear-comment`: Removes the `user.xdg.comment` attribute.
- `--set-rating=N`: Sets rating attribute (e.g. `user.rating` or `user.xdg.rating` based on convention, probably `user.xdg.rating` or `user.rating`? I will use `user.xdg.rating` but keep it generic. Wait, XDG metadata usually doesn't have rating. I'll use `user.xdg.rating` or `user.rating` based on what the rest of the project does. Actually, I will check what convention is used for rating. If none, I'll use `user.rating` or `user.baloo.rating`). Let's use `user.rating` and `user.xdg.rating`? Wait, I will use `user.xdg.rating` to be safe.
- `--clear-rating`: Removes the rating attribute.

Mutation operations will execute and then exit, or execute and then list the paths? Let's just execute them. If paths are passed with mutation flags, it will apply the mutations to those paths. Multiple files can be specified. If multiple mutation flags are provided, they are applied in sequence (e.g., clear tags, then add tags).

## 2. List vs Inspect Mode behavior
- **List Mode**:
  - `--selinux` adds an `SELINUX` column (truncated/compact based on width).
  - `--samba` adds a `DOSATTRIB` column (e.g., `rhsa` for read-only, hidden, system, archive).
  - `--capabilities` adds a `CAPABILITIES` column (marker or compact summary).
  - `--acl` adds an `ACL` column (`+`, `access`, `default`, `both`).
- **Inspect Mode**:
  - Automatically shows SELinux context, Samba DOS attributes, Capabilities, and ACLs grouped under specific sections, in addition to raw xattrs.

## 3. Precedence / Mutual Exclusion
- If mutation flags are used, the tool will perform the mutation and will *not* perform the standard listing scan, or it can perform the mutation and then list. Given standard UNIX tools, it's better to just perform the mutation and maybe output success or nothing, or list the files mutated. I will make mutation flags a distinct mode of operation. If mutation flags are present, we apply them to the given paths and exit.
- `set` and `clear` for the same attribute are mutually exclusive (or `clear` takes precedence before `set`). I'll process `clear` first, then `set`, then `add`, then `remove` to be deterministic. Or error if `set` and `clear` are both provided for the same attribute. Let's error if contradictory flags are provided (e.g., `--set-tags` and `--clear-tags`).

## 4. [ASK] Items
- Should `rating` use `user.rating`, `user.baloo.rating`, or `user.xdg.rating`? (Defaulting to `user.xdg.rating` for consistency with tags/comment, but maybe `user.rating` is more standard?)
