Type: `googlecompute-import`
Artifact BuilderId: `packer.post-processor.googlecompute-import`

The Google Compute Image Import post-processor takes a compressed raw disk
image and imports it to a GCE image available to Google Compute Engine.

~> This post-processor is for advanced users. Please ensure you read the
[GCE import documentation](https://cloud.google.com/compute/docs/images/import-existing-image)
before using this post-processor.

## How Does it Work?

The import process operates by uploading a temporary copy of the compressed raw
disk image to a GCS bucket, and calling an import task in GCP on the raw disk
file. Once completed, a GCE image is created containing the converted virtual
machine. The temporary raw disk image copy in GCS can be discarded after the import is complete.

Google Cloud has very specific requirements for images being imported. Please
see the [GCE import documentation](https://cloud.google.com/compute/docs/images/import-existing-image)
for details.

~> **Note**: To prevent Packer from deleting the compressed RAW disk image set the `keep_input_artifact` configuration option to `true`.
See [Post-Processor Input Artifacts](/packer/docs/templates/legacy_json_templates/post-processors#input-artifacts) for more details.

## Authentication

To authenticate with GCE, this builder supports everything the plugin does.
To get more information on this, refer to the plugin's description page, under
the [authentication](/packer/integrations/hashicorp/googlecompute#authentication) section.

## Configuration

### Required

<!-- Code generated from the comments of the Config struct in post-processor/googlecompute-import/post-processor.go; DO NOT EDIT MANUALLY -->

- `project_id` (string) - The project ID where the GCS bucket exists and where the GCE image is stored.

- `bucket` (string) - The name of the GCS bucket where the raw disk image will be uploaded.

- `image_name` (string) - The unique name of the resulting image.

<!-- End of code generated from the comments of the Config struct in post-processor/googlecompute-import/post-processor.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in post-processor/googlecompute-import/post-processor.go; DO NOT EDIT MANUALLY -->

- `scopes` ([]string) - The service account scopes for launched importer post-processor instance.
  Defaults to:
  
  ```json
  [
    "https://www.googleapis.com/auth/cloud-platform"
  ]
  ```

- `gcs_object_name` (string) - The name of the GCS object in `bucket` where
  the RAW disk image will be copied for import. This is treated as a
  [template engine](/packer/docs/templates/legacy_json_templates/engine). Therefore, you
  may use user variables and template functions in this field. Defaults to
  `packer-import-{{timestamp}}.tar.gz`.

- `image_architecture` (string) - Specifies the architecture or processor type that this image can support. Must be one of: `arm64` or `x86_64`. Defaults to `ARCHITECTURE_UNSPECIFIED`.

- `image_description` (string) - The description of the resulting image.

- `image_family` (string) - The name of the image family to which the resulting image belongs.

- `image_guest_os_features` ([]string) - A list of features to enable on the guest operating system. Applicable only for bootable images. Valid
  values are `MULTI_IP_SUBNET`, `UEFI_COMPATIBLE`,
  `VIRTIO_SCSI_MULTIQUEUE`, `GVNIC` and `WINDOWS` currently.

- `image_labels` (map[string]string) - Key/value pair labels to apply to the created image.

- `image_storage_locations` ([]string) - Specifies a Cloud Storage location, either regional or multi-regional, where image content is to be stored. If not specified, the multi-region location closest to the source is chosen automatically.

- `skip_clean` (bool) - Skip removing the TAR file uploaded to the GCS
  bucket after the import process has completed. "true" means that we should
  leave it in the GCS bucket, "false" means to clean it out. Defaults to
  `false`.

- `image_platform_key` (string) - A key used to establish the trust relationship between the platform owner and the firmware. You may only specify one platform key, and it must be a valid X.509 certificate.

- `image_key_exchange_key` ([]string) - A key used to establish a trust relationship between the firmware and the OS. You may specify multiple comma-separated keys for this value.

- `image_signatures_db` ([]string) - A database of certificates that are trusted and can be used to sign boot files. You may specify single or multiple comma-separated values for this value.

- `image_forbidden_signatures_db` ([]string) - A database of certificates that have been revoked and will cause the system to stop booting if a boot file is signed with one of them. You may specify single or multiple comma-separated values for this value.

<!-- End of code generated from the comments of the Config struct in post-processor/googlecompute-import/post-processor.go; -->


## Basic Example

Here is a basic example. This assumes that the builder has produced an
compressed raw disk image artifact for us to work with, and that the GCS bucket
has been created.

**HCL**

```hcl
post-processor "googlecompute-import"{
  account_file = "account.json"
  bucket = "my-bucket"
  project_id = "my-project"
  image_name = "my-gce-image"
}
```

**JSON**

```json
{
  "type": "googlecompute-import",
  "account_file": "account.json",
  "project_id": "my-project",
  "bucket": "my-bucket",
  "image_name": "my-gce-image"
}
```


## QEMU Builder Example

Here is a complete example for building a Fedora 31 server GCE image. For this
example Packer was run from a Debian Linux host with KVM installed.

    $ packer build -var serial=$(tty) build.pkr.hcl

**HCL2**

```hcl
variables {
  account_file = "account.json"
  bucket = "my-bucket"
  project = "my-project"
  serial = ""
}

source "qemu" "example" {
    accelerator = "kvm"
    boot_command = [
      "<tab> console=ttyS0,115200n8 inst.text inst.ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/fedora-31-ks.cfg rd.live.check=0<enter><wait>"
    ]
    disk_size = "15000"
    format = "raw"
    iso_checksum = "sha256:225ebc160e40bb43c5de28bad9680e3a78a9db40c9e3f4f42f3ee3f10f95dbeb"
    iso_url = "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/31/Server/x86_64/iso/Fedora-Server-dvd-x86_64-31-1.9.iso"
    headless = "true"
    http_directory = "http"
    http_port_max = "10089"
    http_port_min = "10082"
    output_directory = "output"
    shutdown_timeout = "30m"
		shutdown_command = "echo 'vagrant'|sudo -S shutdown -P now"
		ssh_username = "vagrant"
		ssh_password = "vagrant"
    vm_name = "disk.raw"
    qemu_binary = "/usr/bin/kvm"
    qemuargs = [
      ["-m", "1024"],
      ["-cpu", "host"],
      ["-chardev", "tty,id=pts,path=${var.serial}"],
      ["-device", "isa-serial,chardev=pts"],
      ["-device", "virtio-net,netdev=user.0"]
    ]
}

build {
  sources = ["source.qemu.example"]

  post-processors {
    post-processor "compress" {
        output = "output/disk.raw.tar.gz"
    }
    post-processor "googlecompute-import"  {
        account_file = var.account_file
        bucket = var.bucket
        project_id = var.project
        image_name = "fedora31-server-packertest"
        image_description = "Fedora 31 Server"
        image_family = "fedora31-server"
      }
    }
}
```

  **JSON**

```json
{
  "variables": {
    "account_file": "account.json",
    "bucket": "my-bucket",
    "project": "my-project",
    "serial": ""
  },
  "builders": [
    {
      "type": "qemu",
      "accelerator": "kvm",
      "boot_command": [
        "<tab> console=ttyS0,115200n8 inst.text inst.ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/fedora-31-ks.cfg rd.live.check=0<enter><wait>"
      ],
      "disk_size": "15000",
      "format": "raw",
      "iso_checksum": "sha256:225ebc160e40bb43c5de28bad9680e3a78a9db40c9e3f4f42f3ee3f10f95dbeb",
      "iso_url": "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/31/Server/x86_64/iso/Fedora-Server-dvd-x86_64-31-1.9.iso",
      "headless": "true",
      "http_directory": "http",
      "http_port_max": "10089",
      "http_port_min": "10082",
      "output_directory": "output",
      "shutdown_timeout": "30m",
      "shutdown_command": "echo 'vagrant'|sudo -S shutdown -P now",
      "ssh_username": "vagrant",
      "ssh_password": "vagrant",
      "vm_name": "disk.raw",
      "qemu_binary": "/usr/bin/kvm",
      "qemuargs": [
        ["-m", "1024"],
        ["-cpu", "host"],
        ["-chardev", "tty,id=pts,path={{user `serial`}}"],
        ["-device", "isa-serial,chardev=pts"],
        ["-device", "virtio-net,netdev=user.0"]
      ]
    }
  ],
  "post-processors": [
    [
      {
        "type": "compress",
        "output": "output/disk.raw.tar.gz"
      },
      {
        "type": "googlecompute-import",
        "project_id": "{{user `project`}}",
        "account_file": "{{user `account_file`}}",
        "bucket": "{{user `bucket`}}",
        "image_name": "fedora31-server-packertest",
        "image_description": "Fedora 31 Server",
        "image_family": "fedora31-server"
      }
    ]
  ]
}
```
