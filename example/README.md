# Example

## Simple Packer Build

After cloning this repo, move to the `example` directory by running:

```sh
$ cd packer-plugin-scaleway/example
```

Either modify `build_scaleway.pkr.hcl` to reflect your Scaleway keys and project id, or comment that out and set the environment variables by running:

```sh
$ export SCW_DEFAULT_PROJECT_ID=<your scaleway project id>
$ export SCW_ACCESS_KEY=<your scaleway access key>
$ export SCW_SECRET_KEY=<your scaleway secret key>
```

Then run the following commands to build a simple Scaleway image via Packer:

```sh
$ packer init build_scaleway.pkr.hcl
$ packer build build_scaleway.pkr.hcl
```
