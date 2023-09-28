let
  unstable = import (fetchTarball https://nixos.org/channels/nixos-unstable/nixexprs.tar.xz) { };
in
{ nixpkgs ? import <nixpkgs> {} }:
with nixpkgs; mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    gcc
    unstable.go_1_21
    pkg-config
    imagemagick
    pngquant
  ];
}
