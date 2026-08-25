{
  description = "Infuse - manage files across git repositories";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  # Offer prebuilt binaries from the Cachix cache so `nix profile install`
  # downloads instead of compiling. Consumers are prompted to trust these.
  nixConfig = {
    extra-substituters = [ "https://oschrenk.cachix.org" ];
    extra-trusted-public-keys = [
      "oschrenk.cachix.org-1:3JOMfkq2vFiLw4UsCVwzu8kWFBkuS/3DD5AojcO9pks="
    ];
  };

  outputs =
    { self, nixpkgs }:
    let
      # Single source of truth for the version: ./VERSION holds a bare semver
      # (e.g. 0.1.2); the "v" prefix is added by the taskfile release flow and
      # by the ldflags below, so `infuse version` matches the git tag.
      version = nixpkgs.lib.fileContents ./VERSION;

      # aarch64-darwin only: it is what CI builds, so it is the only system the
      # binary cache is ever populated for.
      systems = [
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        infuse = pkgs.buildGoModule {
          pname = "infuse";
          inherit version;
          src = self;

          # Regenerate after changing go.mod/go.sum: set to lib.fakeHash,
          # run `nix build`, then paste the expected hash from the error.
          vendorHash = "sha256-lKhZIgKQPLppAlXHXvD7U/aWJE/hQ5wEuB92Jj8S3To=";

          # Commit and BuildDate keep their "unknown" defaults: the sandbox has
          # no .git and a build date would make the derivation unreproducible.
          ldflags = [
            "-s"
            "-w"
            "-X github.com/oschrenk/infuse/internal/cli.Version=v${version}"
          ];

          # `completion` is hidden (internal/cli/root.go) but not disabled, and
          # it never reaches the config loader, so it is safe in the sandbox.
          nativeBuildInputs = [ pkgs.installShellFiles ];
          postInstall = ''
            installShellCompletion --cmd infuse \
              --bash <($out/bin/infuse completion bash) \
              --zsh <($out/bin/infuse completion zsh) \
              --fish <($out/bin/infuse completion fish)
          '';

          meta = {
            description = "Manage files across git repositories";
            homepage = "https://github.com/oschrenk/infuse";
            mainProgram = "infuse";
            platforms = nixpkgs.lib.platforms.darwin;
          };
        };
        default = infuse;
      });

      apps = forAllSystems (pkgs: rec {
        infuse = {
          type = "app";
          program = "${self.packages.${pkgs.stdenv.hostPlatform.system}.infuse}/bin/infuse";
        };
        default = infuse;
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go # go, language
            go-task # go, task runner
            gopls # go, lsp
            gotools # go, goimports
            go-tools # go, staticcheck
          ];
        };
      });
    };
}
