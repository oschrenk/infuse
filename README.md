# README

Infuse files into Git repositories

## Usecases

### Tracking TODOs

- you want to have a `TODO.md` in the repository
- you don't want the `TODO.md` file to be part of the git history
- it shouldn't appear in `git status`
- but you still want to track changes in the `TODO.md`

### Tracking CLAUDE.md

- you want to have a `CLAUDE.md` in the repository
- but don't want colleagues to know
- but still want to track changes over time

## Commands

```
# if there is a config, it should do the thing
# otherwise offer to create global config
infuse

# the source repo
# --auto-commit
# --auto-push

# create

infuse config        # config repo for infusion

# --no-exclude
# --with-exclude // default
#
# I want to be able track these changes in a different repo

# moves file to configured repo
infuse track path/to/file  # track file
infuse clean           # clean all infused changes

# operate on infuse repo
infuse git *           # forward * to infuse repo
```

## Configuration

These paths are checked in order

- `$XDG_CONFIG_PATH/infuse/config.toml`
- `$HOME/infuse/config.toml`

Example configuration

```
[repo]
path = "$XDG_DATA_HOME/infuse"
"
```

These fule
-






- data repo should be created in $HOME/.local/share/infuse
- git repository


