{
  description = "libations - a web app for viewing cocktail recipes";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { self
    , nixpkgs
    , flake-utils
    , ...
    }:
    flake-utils.lib.eachSystem [
      "x86_64-linux"
      "aarch64-linux"
    ]
      (system:
      let
        pkgs = import nixpkgs { inherit system; };
        # Generate a user-friendly version number.
        version = builtins.substring 0 8 self.lastModifiedDate;
      in
      {
        packages = rec {
          libations = pkgs.buildGoModule {
            pname = "libations";
            inherit version;
            src = ./.;
            vendorHash = "sha256-Ep3nBl9WZm7skk1cmMS9KI019ZSRSxofbLs2Nrj6HM8=";
            nativeBuildInputs = with pkgs; [ hugo ];
            postConfigure = "go generate";
          };
          default = libations;
        };

        devShells = {
          default = pkgs.mkShell {
            name = "libations";
            NIX_CONFIG = "experimental-features = nix-command flakes";
            nativeBuildInputs = with pkgs; [
              go_1_21
              go-tools
              gofumpt
              gopls
              hugo
            ];
            shellHook = "exec $SHELL";
          };
        };
      }
      );
}
