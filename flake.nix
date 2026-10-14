{
  description = "Go devshell with air hot reload";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gotools

            pnpm
            nodejs_24
          ];

          packages = with pkgs; [
            air
            go-swag
          ];

          shellHook = ''
            echo "🚀 Go devshell ready!"
            echo "Run: air (hot reload) or go run main.go"
          '';

          env = {
            GO111MODULE = "on";
          };
        };
      });
}