{ pkgs ? import <nixpkgs> {} }:

let
  # Fetch Hugo v0.160.1 binary (tar.gz)
  hugo-tarball = pkgs.fetchurl {
    url = "https://github.com/gohugoio/hugo/releases/download/v0.160.1/hugo_0.160.1_linux-amd64.tar.gz";
    sha256 = "sha256:2f72fefa6ce944907f74c36c35abcaa306cdd22e64647a9ee5b328fc0bfb67be";
  };

  # Extract Hugo binary from the tarball
  hugo-bin = pkgs.runCommand "hugo-bin" {
    nativeBuildInputs = [ pkgs.btar ];
  } ''
    mkdir -p $out/bin
    tar -xzf ${hugo-tarball} -C $out/bin
    chmod +x $out/bin/hugo
  '';

in

pkgs.mkShell {
  name = "hugo-tailwind-env";

  buildInputs = with pkgs; [
    hugo-bin
    # Use the prepackaged version, because nixOS cannot run dynamically linked executables
    tailwindcss_4
    # I need to convert large photo, in something that loads fast in the browser
    libwebp
  ];

  shellHook = ''
    echo "Willkommen in der Entwicklungsumgebung von Elses Website!"
    # tailwindcss --version
    hugo version
    cwebp -version
  '';
}
