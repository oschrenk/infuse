# infuse

Manage files across git repositories.

## Motivation

- You want files like `CLAUDE.md` available in your repository
- You don't want to commit these files to the repository
- You want to track changes to these files in a central location

## Getting started

```
infuse init          # create config file
# edit ~/.config/infuse/config.toml to set your repo path
infuse setup         # prepare working repository
```

## Commands

### `infuse init`

Create the config file at `$XDG_CONFIG_HOME/infuse/config.toml`.

### `infuse setup`

Prepare the current working repository for use with infuse. Creates the infuse repository and directory structure if needed.

### `infuse add`

Move files into the infuse repository and symlink them back.

```
infuse add CLAUDE.md
infuse add src/config/settings.yaml
infuse add file1 file2 file3
```

### `infuse dir`

Print the infuse repository path. Useful with:

```
cd $(infuse dir)
```

### `infuse status`

Show all files managed by infuse for the current working repository.

```
$ infuse status
PATH              SYMLINK  GIT
CLAUDE.md         ok       clean
taskfile.yml      ok       modified
.env              missing  clean
```

## Configuration

Config file location: `$XDG_CONFIG_HOME/infuse/config.toml`

Example:

```toml
[repo]
path = "$XDG_DATA_HOME/infuse"
```

The infuse repository must be a git repository. Run `infuse setup` to initialize it.
