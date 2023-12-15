{ buildGoModule
, hugo
, lastModifiedDate
, lib
, ...
}:

let
  version = builtins.substring 0 8 lastModifiedDate;
in
buildGoModule {
  pname = "libations";
  inherit version;
  src = lib.cleanSource ../.;
  vendorHash = "sha256-Ep3nBl9WZm7skk1cmMS9KI019ZSRSxofbLs2Nrj6HM8=";
  nativeBuildInputs = [ hugo ];
  postConfigure = "go generate";
}
