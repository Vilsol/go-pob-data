{ nixpkgs ? import <nixpkgs> {} }:
with nixpkgs; mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    gcc
    go
    pkg-config
    imagemagick
    pngquant
  ];
}
