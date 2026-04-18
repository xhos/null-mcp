{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    git-hooks.url = "github:cachix/git-hooks.nix";
    git-hooks.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = {
    self,
    nixpkgs,
    git-hooks,
  }: let
    forAllSystems = f:
      nixpkgs.lib.genAttrs
      ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"]
      (system: f nixpkgs.legacyPackages.${system});
  in {
    checks = forAllSystems (pkgs: {
      pre-commit = git-hooks.lib.${pkgs.system}.run {
        src = ./.;
        hooks = {
          govet.enable = true;
          alejandra.enable = true;
          golangci-lint = {
            enable = true;
            name = "golangci-lint";
            entry = "${pkgs.golangci-lint}/bin/golangci-lint fmt";
            types = ["go"];
          };
        };
      };
    });

    packages = forAllSystems (pkgs: {
      default = pkgs.buildGoModule {
        pname = "null-mcp";
        version = self.shortRev or self.dirtyShortRev or "dev";

        src = ./.;

        vendorHash = "sha256-If9j2lxrZPatKK11Lc670kC0ypjE8uL9iTyCOXVTVSc=";

        ldflags = let
          pkg = "github.com/xhos/null-mcp/internal/version";
          ver = self.shortRev or self.dirtyShortRev or "dev";
        in [
          "-X ${pkg}.Version=${ver}"
          "-X ${pkg}.GitCommit=${self.shortRev or "dirty"}"
          "-X ${pkg}.GitBranch=main"
          "-X ${pkg}.BuildTime=${toString self.lastModified}"
        ];
      };
    });

    devShells = forAllSystems (pkgs: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          go
          golangci-lint
          air

          buf
          protoc-gen-go
          protoc-gen-go-grpc
          protoc-gen-connect-go

          (writeShellScriptBin "run" ''
            exec ${air}/bin/air -build.cmd "go build -o ./tmp/main ./cmd/main.go" -build.bin ./tmp/main
          '')

          (writeShellScriptBin "regen" ''
            rm -rf internal/gen/
            ${buf}/bin/buf generate
          '')

          (writeShellScriptBin "fmt" ''
            ${golangci-lint}/bin/golangci-lint fmt
          '')

          (writeShellScriptBin "tst" ''
            go test ./...
          '')

          (writeShellScriptBin "tstv" ''
            CLICOLOR_FORCE=1 go test ./... -v
          '')

          (writeShellScriptBin "bump-protos" ''
            git -C proto fetch origin
            git -C proto checkout main
            git -C proto pull --ff-only
            git add proto
            git commit -m "chore: bump proto files"
            git push
          '')
        ];

        shellHook = "${self.checks.${pkgs.system}.pre-commit.shellHook}";
      };
    });
  };
}
