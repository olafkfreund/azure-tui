{
  description = "Azure TUI/CLI Go application with AI, TUI, and Azure SDK integration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
        goPkgs = with pkgs; [
          go_1_22
          gopls
          gotools
          go-tools
          golangci-lint
          just
          git
          # TUI/AI dependencies
          nodejs_20
          # Optional: for OpenAI CLI testing
          curl
        ];
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = goPkgs;
          shellHook = ''
            export GOPATH=$PWD/.gopath
            export GOBIN=$PWD/.gopath/bin
            export PATH=$GOBIN:$PATH
            export GO111MODULE=on
            echo "Azure TUI dev shell loaded. Run 'just' for build/test commands."
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "azure-tui";
          version = "0.1.0";
          src = ./.;
          vendorSha256 = null;
          subPackages = ["cmd"];
          doCheck = false;
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };
      }
    );
}
