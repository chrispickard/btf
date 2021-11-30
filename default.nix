{ pkgs ? import <nixpkgs> { }, stdenv ? pkgs.stdenv, lib ? pkgs.lib
, buildGoModule ? pkgs.buildGoModule, fetchFromGitHub ? pkgs.fetchFromGitHub
  # We use this to add matchers for stuff that's not in upstream nixpkgs, but is
  # in our own overlay. No fuzzy matching from multiple options here, it's just:
  # Was the command `, mything`? Run `nixpkgs.mything`.
, overlayPackages ? [ ] }:

buildGoModule rec {
  pname = "btf";
  version = lib.removeSuffix "\n" (builtins.readFile ./VERSION);

  src = ./.;

  vendorSha256 = "1pdp7a43lw0jzqsca63c501ra659l0231zjkydi69632zghc80as";

  # Since the tarball pulled from GitHub doesn't contain git tag information,
  # we fetch the expected tag's timestamp from a file in the root of the
  # repository.
  preBuild = ''
    buildFlagsArray=(
      -ldflags="
        -X github.com/chrispickard/btf/version.VERSION=${version}
      "
    )
  '';

  meta = with lib; {
    homepage = "https://github.com/chrispickard/btf";
    description = "A simple, keyboard driven app switcher/launcher for x11";
    license = licenses.asl20;
  };
}
