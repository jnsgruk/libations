{ config, lib, pkgs, ... }:
let
  cfg = config.services.libations;
in
{
  options = {
    services.libations = {
      enable = lib.mkEnableOption "Enables the libations service";

      recipesFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        example = "/var/lib/libations/recipes.json";
        description = lib.mdDoc ''
          A file containing drinks recipes per the Libations file format.
          See https://github.com/jnsgruk/libations.
        '';
      };

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
        ExecStart = "${cfg.package}/bin/libations -recipes-file ${cfg.recipesFile}";
        Restart = "always";
        EnvironmentFile = cfg.tailscaleKeyFile;
        StateDirectory = "libations";
        StateDirectoryMode = "0750";
      };
    };
  };
}
