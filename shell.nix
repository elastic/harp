{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [
    go_1_17
    gotools
    gopls
    go-outline
    gocode
    gopkgs
    gocode-gomod
    godef
    golint
    delve
    mage
    protobuf
    golangci-lint
    upx
  ];
  shellHook =
  ''
    export PATH=$(pwd)/tools/bin:$PATH
  '';
}
