The Google compute Packer plugin lets you create custom images for use within Google Compute Engine (GCE).

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    googlecompute = {
      source  = "github.com/hashicorp/googlecompute"
      version = "~> 1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/hashicorp/googlecompute
```

### Components

#### Builders

- [googlecompute](/packer/integrations/hashicorp/googlecompute/latest/components/builder/googlecompute) - The
  googlecompute builder creates images from existing ones, by launching an instance, provisioning it, then exporting
  it as a reusable image.

#### Post-Processors

- [googlecompute-import](/packer/integrations/hashicorp/googlecompute/latest/components/post-processor/googlecompute-import) -
  The googlecompute-import post-processor imports an existing raw disk image, and imports it as a GCE image that can be
  used for launching instances from.

- [googlecompute-export](/packer/integrations/hashicorp/googlecompute/latest/components/post-processor/googlecompute-export) -
  The googlecompute-export post-processor exports the image built by the googlecompute builder as a .tar.gz archive into Google
  Cloud Storage (GCS).
