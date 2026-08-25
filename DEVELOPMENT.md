# Development

## Requirements

- [Go](https://go.dev/) — `brew install go`
- [Task](https://taskfile.dev/) — `brew install go-task`
- [staticcheck](https://staticcheck.dev/) — `go install honnef.co/go/tools/cmd/staticcheck@latest`
- [gh](https://cli.github.com/) — `brew install gh` (for releases)
- [nix](https://nixos.org/) — optional, provides the dev shell and the flake build
- [direnv](https://direnv.net/) — optional, loads the dev shell on `cd`

With nix and direnv installed, `direnv allow` enters the dev shell from `.envrc`, which pins go, go-task, gopls and staticcheck.

## Commands

- `task artifacts` — Build release artifacts to `.release/`
- `task build` — Build binary
- `task clean` — Remove build artifacts
- `task install` — Install to `$GOPATH/bin/` and fish completion
- `task lint` — Lint (staticcheck, tidy, vet)
- `task release` — Create GitHub release with artifacts from `VERSION`
- `task run` — Run from source
- `task sha` — Generate SHA256 checksums
- `task tag` — Create and push git tag from `VERSION`
- `task test` — Run tests
- `task uninstall` — Remove from `$GOPATH/bin/` and fish completion
- `task updates` — Check for dependency updates
- `task version` — Show current version info

## Nix

`VERSION` holds a bare semver and is the single source of truth.
The taskfile adds the `v` prefix for tags and ldflags, and `flake.nix` reads the file as-is.

- `nix build .#infuse` — build the package
- `nix develop` — enter the dev shell
- `nix run .#infuse -- status` — run without installing

Changing `go.mod` or `go.sum` invalidates `vendorHash` in `flake.nix`.
Set it to `lib.fakeHash`, run `nix build`, and paste the expected hash from the error.

## Release

1. Ensure all changes are committed and pushed
2. Bump the version in `VERSION`
3. `task release` to build the artifacts and publish the GitHub release
4. `task sha` to get hashes
5. Update [homebrew-made](https://github.com/oschrenk/homebrew-made) with new version and SHA
6. `brew update && brew upgrade infuse`
