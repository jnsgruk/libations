{
  buildGo122Module,
  lastModifiedDate,
  lib,
  ...
}:

let
  version = builtins.substring 0 8 lastModifiedDate;
in
buildGo122Module {
  pname = "libations";
  inherit version;
  src = lib.cleanSource ../.;
  vendorHash = "sha256-AWvaHyJL7Cm+zCY/vTuTAsgLbVy6WUNfmaGbyQOzMMQ=";
}
