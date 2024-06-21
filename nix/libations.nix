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
  vendorHash = "sha256-qnWiByRitZ6rvB0zbDo6Jhh5pXBHz3/+IY5fCoBAdrE=";
}
