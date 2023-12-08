{
  description = "libations - a web app for viewing cocktail recipes";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs =
    { self
    , nixpkgs
    , ...
    }:
    let
      forAllSystems = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "aarch64-linux"
      ];

      pkgsForSystem = system: (import nixpkgs {
        inherit system;
        overlays = [ self.overlays.default ];
      });

      version = builtins.substring 0 8 self.lastModifiedDate;
    in
    {
      overlays.default = _final: _prev: {
        libations = self.packages.libations;
      };

      packages = forAllSystems (system:
        let
          pkgs = pkgsForSystem system;
        in
        rec {
          libations = pkgs.buildGoModule {
            pname = "libations";
            inherit version;
            src = ./.;
            vendorHash = "sha256-Ep3nBl9WZm7skk1cmMS9KI019ZSRSxofbLs2Nrj6HM8=";
            nativeBuildInputs = with pkgs; [ hugo ];
            postConfigure = "go generate";
          };
          default = libations;
        });

      nixosModules = rec {
        default = libations;
        libations = { config, lib, pkgs, ... }:
          let
            cfg = config.services.libations;
          in
          {
            options = {
              services.libations = {
                enable = lib.mkEnableOption "Enables the libations service";

                tailscaleKeyFile = lib.mkOption {
                  type = lib.types.nullOr lib.types.path;
                  default = null;
                  example = "/run/agenix/libations-tsauthkey";
                  description = lib.mdDoc ''
                    A file containing a key for Libations to join a Tailscale network.
                    See https://tailscale.com/kb/1085/auth-keys/.
                  '';
                };

                package = lib.mkPackageOptionMD pkgs "libations" { };
              };
            };

            config = lib.mkIf cfg.enable {
              systemd.services.libations = {
                description = "Libations cocktail recipe viewer";
                wantedBy = [ "multi-user.target" ];
                after = [ "network.target" ];
                environment = {
                  "XDG_CONFIG_HOME" = "/var/lib/libations/";
                };
                serviceConfig = {
                  DynamicUser = true;
                  ExecStart = "${cfg.package}/bin/libations";
                  Restart = "always";
                  EnvironmentFile = cfg.tailscaleKeyFile;
                  StateDirectory = "libations";
                  StateDirectoryMode = "0750";
                };
              };
            };
          };


        devShells = forAllSystems (system:
          let
            pkgs = pkgsForSystem system;
          in
          {
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
          });
      };
    };
}

