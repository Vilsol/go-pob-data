{ nixpkgs ? import <nixpkgs> {} }:
with nixpkgs; mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    gcc
    pkg-config
    pngquant
  ];
  buildInputs = [
    imagemagick
  ];
}
