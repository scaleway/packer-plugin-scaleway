packer {
  required_plugins {
    scaleway = {
      version = ">= 1.0.5"
      source  = "github.com/scaleway/scaleway"
    }
  }
}

source "scaleway" "basic" {
  project_id = "YOUR PROJECT ID"
  access_key = "YOUR ACCESS KEY"
  secret_key = "YOUR SECRET KEY"
  commercial_type = "DEV1-S"
  image = "ubuntu_focal"
  image_name = "basic build"
  ssh_username = "root"
  zone = "fr-par-1"
  tags = ["foo", "bar" ]
}

build {
  sources = ["source.scaleway.basic"]
}
