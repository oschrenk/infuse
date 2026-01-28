# Spec

## Definitions

- **Working repository**: The git repository the user is currently working in. Files are symlinked here.
- **Infuse repository** (`$INFUSE_REPO`): The central git repository that stores tracked files, organized by `<host>/<owner>/<repo>`.

## Commands

- `infuse init` — initialize infuse configuration
- `infuse setup` — prepare a working repository for use with infuse
- `infuse add` — move a file into the infuse repository and symlink it back
- `infuse status` — show managed files for the current working repository

### `infuse init`

First-time setup. Creates the config file at `$XDG_CONFIG_HOME/infuse/config.toml` with default contents.

### `infuse setup`

Prepares a working repository for use with infuse.

Preconditions:
- Configuration must exist and have valid syntax
- `$INFUSE_REPO` path must be configured

Steps:
1. If `$INFUSE_REPO` directory does not exist, prompt the user to create it
2. If `$INFUSE_REPO` is not a git repository, prompt the user to initialize it
3. Verify the working directory is inside a git repository
4. Verify the git repository has a remote configured
5. Verify the working repository is not the infuse repository (no recursion)
6. Extract `<host>/<owner>/<repo>` from the remote URL
7. Create the directory structure inside `$INFUSE_REPO`

### `infuse add`

Moves a file from the working repository into the infuse repository and creates a symlink back.

Preconditions:
- Configuration must exist and be valid
  - If not, fail with an error message to run `infuse init`
- Working repository must be a git repository with a remote configured
- `$INFUSE_REPO` must be a git repository
  - If not, fail with an error message to run `infuse setup`
- `$INFUSE_REPO` must contain the `<host>/<owner>/<repo>` directory
- The target path
  - must exist in the working repository
  - must not be tracked by git in the working repository
  - must not be a symlink
  - may already be tracked by infuse (only untracked paths will be added)

Steps:
1. Move the file to `$INFUSE_REPO/<host>/<owner>/<repo>/...`, preserving subdirectory structure
2. Create a symlink from the original location to the new location in `$INFUSE_REPO`
3. Add an entry to `.git/info/exclude`

Example:

```
mkdir -p .git/info/
echo taskfile.yml >> .git/info/exclude
ln -s /Users/oliver/.local/share/infuse/git.timewax.com/timewax/backend/taskfile.yml taskfile.yml
```

Flags:
- `--auto-commit` — automatically commit the change in the infuse repository
- `--message` — commit message
- `--push` — push after committing

Usage:

```
infuse add path/to/file
```

### `infuse status`

Shows all files managed by infuse for the current working repository.

Preconditions:
- Configuration must exist and be valid
- Must be run from inside a git repository with a remote configured
- `$INFUSE_REPO` must contain the `<host>/<owner>/<repo>` directory

Output per file:
- Path relative to the working repository root
- Symlink status: whether the symlink exists and is valid, or is broken/missing
- Git status: whether the file has uncommitted changes in the infuse repository
