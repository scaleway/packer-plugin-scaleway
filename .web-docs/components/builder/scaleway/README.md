Type: `scaleway`
Artifact BuilderId: `hashicorp.scaleway`

The `scaleway` Packer builder is able to create new images for use with
[Scaleway](https://www.scaleway.com). The builder takes a source image, runs
any provisioning necessary on the image after launching it, then snapshots it
into a reusable image. This reusable image can then be used as the foundation
of new servers that are launched within Scaleway.

The builder does _not_ manage snapshots. Once it creates an image, it is up to
you to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/legacy_json_templates/communicator) can be configured for this
builder. In addition to the options defined there, a private key file
can also be supplied to override the typical auto-generated key:

- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


### Required:

<!-- Code generated from the comments of the Config struct in builder/scaleway/config.go; DO NOT EDIT MANUALLY -->

- `access_key` (string) - The AccessKey corresponding to the secret key.
  Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
  It can also be specified via the environment variable SCW_ACCESS_KEY.

- `secret_key` (string) - The SecretKey to authenticate against the Scaleway API.
  Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
  It can also be specified via the environment variable SCW_SECRET_KEY.

- `project_id` (string) - The Project ID in which the instances, volumes and snapshots will be created.
  Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
  It can also be specified via the environment variable SCW_DEFAULT_PROJECT_ID.

- `zone` (string) - The Zone in which the instances, volumes and snapshots will be created.
  Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
  It can also be specified via the environment variable SCW_DEFAULT_ZONE

- `image` (string) - The UUID of the base image to use. This is the image
  that will be used to launch a new server and provision it. See
  the images list
  get the complete list of the accepted image UUID.
  The marketplace image label (eg `ubuntu_focal`) also works.

- `commercial_type` (string) - The name of the server commercial type:
  DEV1-S, DEV1-M, DEV1-L, DEV1-XL,
  PLAY2-PICO, PLAY2-NANO, PLAY2-MICRO,
  PRO2-XXS, PRO2-XS, PRO2-S, PRO2-M, PRO2-L,
  GP1-XS, GP1-S, GP1-M, GP1-L, GP1-XL,
  ENT1-XXS, ENT1-XS, ENT1-S, ENT1-M, ENT1-L, ENT1-XL, ENT1-2XL,
  GPU-3070-S, RENDER-S, STARDUST1-S,

<!-- End of code generated from the comments of the Config struct in builder/scaleway/config.go; -->


### Optional:

<!-- Code generated from the comments of the Config struct in builder/scaleway/config.go; DO NOT EDIT MANUALLY -->

- `api_url` (string) - The Scaleway API URL to use
  Will be fetched first from the [scaleway configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
  It can also be specified via the environment variable SCW_API_URL

- `image_size_in_gb` (int32) - The Image size in GB. Will only work for images based on block volumes.

- `snapshot_name` (string) - The name of the resulting snapshot that will
  appear in your account. Default packer-TIMESTAMP

- `image_name` (string) - The name of the resulting image that will appear in
  your account. Default packer-TIMESTAMP

- `server_name` (string) - The name assigned to the server. Default
  packer-UUID

- `bootscript` (string) - The id of an existing bootscript to use when
  booting the server.

- `boottype` (string) - The type of boot, can be either local or
  bootscript, Default bootscript

- `remove_volume` (bool) - RemoveVolume remove the temporary volumes created before running the server

- `block_volume` ([]ConfigBlockVolume) - BlockVolumes define block volumes attached to the server alongside the default volume
  See the [BlockVolumes](#block-volumes-configuration) documentation for fields.

- `cleanup_machine_related_data` (string) - This value allows the user to remove information
  that is particular to the instance used to build the image

- `snapshot_creation_timeout` (duration string | ex: "1h5m2s") - The time to wait for snapshot creation. Defaults to "1h"

- `image_creation_timeout` (duration string | ex: "1h5m2s") - The time to wait for image creation. Defaults to "1h"

- `server_creation_timeout` (duration string | ex: "1h5m2s") - The time to wait for server creation. Defaults to "10m"

- `server_shutdown_timeout` (duration string | ex: "1h5m2s") - The time to wait for server shutdown. Defaults to "10m"

- `user_data` (map[string]string) - User data to apply when launching the instance

- `user_data_timeout` (duration string | ex: "1h5m2s") - A custom timeout for user data to assure its completion. Defaults to "0s"

- `tags` ([]string) - A list of tags to apply on the created image, volumes, and snapshots

- `api_token` (string) - The token to use to authenticate with your account.
  It can also be specified via environment variable SCALEWAY_API_TOKEN. You
  can see and generate tokens in the "Credentials"
  section of the control panel.
  Deprecated: use SecretKey instead

- `organization_id` (string) - The organization id to use to identify your
  organization. It can also be specified via environment variable
  SCALEWAY_ORGANIZATION. Your organization id is available in the
  "Account" section of the
  control panel.
  Previously named: api_access_key with environment variable: SCALEWAY_API_ACCESS_KEY
  Deprecated: use ProjectID instead

- `region` (string) - The name of the region to launch the server in (par1
  or ams1). Consequently, this is the region where the snapshot will be
  available.
  Deprecated: use Zone instead

<!-- End of code generated from the comments of the Config struct in builder/scaleway/config.go; -->


### Block volumes configuration

<!-- Code generated from the comments of the ConfigBlockVolume struct in builder/scaleway/config.go; DO NOT EDIT MANUALLY -->

- `name` (string) - The name of the created volume

- `snapshot_id` (string) - ID of the snapshot to create the volume from

- `size` (uint64) - Size of the newly created volume

<!-- End of code generated from the comments of the ConfigBlockVolume struct in builder/scaleway/config.go; -->


## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

**HCL2**

```hcl
source "scaleway" "example" {
  project_id = "YOUR PROJECT ID"
  access_key = "YOUR ACCESS KEY"
  secret_key = "YOUR SECRET KEY"
  image = "UUID OF THE BASE IMAGE"
  zone = "fr-par-1"
  commercial_type = "DEV1-S"
  ssh_username = "root"
  ssh_private_key_file = "~/.ssh/id_rsa"
}

build {
  sources = ["source.scaleway.example"]
}
```


**JSON**

    ```json
    {
        "type": "scaleway",
        "project_id": "YOUR PROJECT ID",
        "access_key": "YOUR ACCESS KEY",
        "secret_key": "YOUR SECRET KEY",
        "image": "UUID OF THE BASE IMAGE",
        "zone": "fr-par-1",
        "commercial_type": "DEV1-S",
        "ssh_username": "root",
        "ssh_private_key_file": "~/.ssh/id_rsa"
    }
    ```


When you do not specify the `ssh_private_key_file`, a temporary SSH keypair
is generated to connect the server. This key will only allow the `root` user to
connect the server.
