{
  description = "xbom - Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go_1_25

            # Build essentials
            gcc
            pkg-config
            gnumake

            # SSL/TLS libraries
            openssl
            openssl.dev

            # Git hooks manager
            lefthook

            # Required for pre-commit hook
            gitleaks

            # Go linting
            golangci-lint

            # Additional development tools
            git
          ];

          # Environment variables
          CGO_ENABLED = "1";
          GOEXPERIMENT = "greenteagc";

          # Setup instructions and environment
          shellHook = ''
            echo "🔧 xBom development environment"
            echo ""
            echo "Available commands:"
            echo "  make              - Build the project"
            echo "  make clean        - Clean build artifacts"
            echo "  lefthook install  - Install git hooks"
            echo ""
            echo "Go version: $(go version)"
            echo "golangci-lint: $(golangci-lint --version 2>/dev/null || echo 'not found')"
            echo "lefthook: $(lefthook version 2>/dev/null || echo 'not found')"
            echo ""

            # Ensure bin directory exists
            mkdir -p bin
          '';
        };
      }
    );
}
