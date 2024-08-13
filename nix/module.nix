{
  config,
  lib,
  pkgs,
  ...
}:
let
  inherit (lib.modules) mkIf;
  inherit (lib.options) mkEnableOption mkPackageOption mkOption;
  inherit (lib.types) nullOr path;

  cfg = config.services.libations;
in
{
  options = {
    services.libations = {
      enable = mkEnableOption "Enables the libations service";

      recipesFile = mkOption {
        type = nullOr path;
        default = null;
        example = "/var/lib/libations/recipes.json";
        description = ''
          A file containing drinks recipes per the Libations file format.
          See https://github.com/jnsgruk/libations.
        '';
      };

      tailscaleKeyFile = mkOption {
        type = nullOr path;
        default = null;
        example = "/run/agenix/libations-tsauthkey";
        description = ''
          A file containing a key for Libations to join a Tailscale network.
          See https://tailscale.com/kb/1085/auth-keys/.
        '';
      };

      package = mkPackageOption pkgs "libations" { };
    };
  };

  config = mkIf cfg.enable {
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
