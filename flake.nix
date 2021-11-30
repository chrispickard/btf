{
  description = "A simple, keyboard driven app switcher/launcher for x11";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in {

      packages.x86_64-linux.btf = (pkgs.callPackage ./default.nix { });

      defaultPackage.x86_64-linux = self.packages.x86_64-linux.btf;

    };
}
