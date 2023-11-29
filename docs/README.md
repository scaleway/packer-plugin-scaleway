The [Scaleway](https://www.scaleway.com) Packer plugin provides a builder for building images in
Scaleway.

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    scaleway = {
      version = ">= 1.0.5"
      source  = "github.com/scaleway/scaleway"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/scaleway/scaleway v1.0.x
```

This command will install the most recent compatible Scaleway Packer plugin matching
version constraint. If the version constraint is omitted, the most recent
version of the plugin will be installed.

### Components

#### Builders

- [scaleway](/packer/integrations/scaleway/scaleway/latest/components/builder/scaleway) - The Scaleway Packer builder is able to create new images for use with Scaleway Compute Instance servers. 
The builder takes a source image, runs any provisioning necessary on the image after launching it, then snapshots it into a reusable image. 
This reusable image can then be used as the foundation of new servers that are launched within Scaleway.

