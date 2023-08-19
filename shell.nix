{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    gcc
    pkg-config
    imagemagick
    pngquant
  ];
}
