packer {
  required_plugins {
  }
}

source "scaleway" "basic" {
  commercial_type = "PLAY2-PICO"
  zone = "fr-par-2"
  image = "ubuntu_jammy"
  image_name = "packer-e2e-complete"
  ssh_username = "root"
  remove_volume = true

  image_size_in_gb = 42
  snapshot_name = "packer-e2e-complete-snapshot"
}

build {
  sources = ["source.scaleway.basic"]
}
