# Debug

The recommended way to debug a packer plugin is adding prints and running with logs enabled.
As written in [packer debug documentation](https://developer.hashicorp.com/packer/docs/debugging), you can use `PACKER_LOG=1` to enable verbose logging.

## Testing local build

To test a local build, you need to have the plugin in the same folder where you will run packer:


```
> ls
packer-plugin-scaleway build_scaleway.pkr.hcl
> PACKER_LOG=DEBUG SCW_DEBUG=1 packer build build_scaleway.pkr.hcl
...
```
