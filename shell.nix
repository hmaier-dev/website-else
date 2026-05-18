{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  name = "hugo-v0.161.1-env";

  buildInputs = with pkgs; [
    git
    hugo
    nodejs_24
  ];
  shellHook = ''
    npm install tailwindcss
  '';
}

