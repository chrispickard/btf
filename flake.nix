{
  description = "A very basic flake";

  outputs = { self, nixpkgs }: {

    packages.x86_64-linux.btf = import ./default.nix;

    defaultPackage.x86_64-linux = self.packages.x86_64-linux.btf;

  };
}
