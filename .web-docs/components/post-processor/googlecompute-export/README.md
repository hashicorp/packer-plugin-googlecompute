Type: `googlecompute-export`
Artifact BuilderId: `packer.post-processor.googlecompute-export`

The Google Compute Image Exporter post-processor exports the resultant image
from a googlecompute build as a gzipped tarball to Google Cloud Storage (GCS).

The exporter uses the same Google Cloud Platform (GCP) project and
authentication credentials as the googlecompute build that produced the image.
A temporary VM is started in the GCP project using these credentials. The VM
mounts the built image as a disk then dumps, compresses, and tars the image.
The VM then uploads the tarball to the provided GCS `paths` using the same
credentials.

As such, the authentication credentials that built the image must have write
permissions to the GCS `paths`.

~> **Note**: By default the GCE image being exported will be deleted once the image has been exported.
To prevent Packer from deleting the image set the `keep_input_artifact` configuration option to `true`. See [Post-Processor Input Artifacts](/packer/docs/templates/legacy_json_templates/post-processors#input-artifacts) for more details.

## Authentication

To authenticate with GCE, this builder supports everything the plugin does.
To get more information on this, refer to the plugin's description page, under
the [authentication](/packer/integrations/hashicorp/googlecompute#authentication) section.

## Configuration

### Required

<!-- Code generated from the comments of the Config struct in post-processor/googlecompute-export/post-processor.go; DO NOT EDIT MANUALLY -->

- `paths` ([]string) - A list of GCS paths where the image will be exported.
  For example `'gs://mybucket/path/to/file.tar.gz'`

<!-- End of code generated from the comments of the Config struct in post-processor/googlecompute-export/post-processor.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in post-processor/googlecompute-export/post-processor.go; DO NOT EDIT MANUALLY -->

- `scopes` ([]string) - The service account scopes for launched exporter post-processor instance.
  Defaults to:
  
  ```json
  [
    "https://www.googleapis.com/auth/cloud-platform"
  ]
  ```

- `disk_size` (int64) - The size of the export instances disk.
  The disk is unused for the export but a larger size will increase `pd-ssd` read speed.
  This defaults to `200`, which is 200GB.

- `disk_type` (string) - Type of disk used to back the export instance, like
  `pd-ssd` or `pd-standard`. Defaults to `pd-ssd`.

- `machine_type` (string) - The export instance machine type. Defaults to `"n1-highcpu-4"`.

- `network` (string) - The Google Compute network id or URL to use for the export instance.
  Defaults to `"default"`. If the value is not a URL, it
  will be interpolated to `projects/((builder_project_id))/global/networks/((network))`.
  This value is not required if a `subnet` is specified.

- `subnetwork` (string) - The Google Compute subnetwork id or URL to use for
  the export instance. Only required if the `network` has been created with
  custom subnetting. Note, the region of the subnetwork must match the
  `zone` in which the VM is launched. If the value is not a URL,
  it will be interpolated to
  `projects/((builder_project_id))/regions/((region))/subnetworks/((subnetwork))`

- `zone` (string) - The zone in which to launch the export instance. Defaults
  to `googlecompute` builder zone. Example: `"us-central1-a"`

- `service_account_email` (string) - Service Account Email

<!-- End of code generated from the comments of the Config struct in post-processor/googlecompute-export/post-processor.go; -->


## Basic Example

The following example builds a GCE image in the project, `my-project`, with an
account whose keyfile is `account.json`. After the image build, a temporary VM
will be created to export the image as a gzipped tarball to
`gs://mybucket1/path/to/file1.tar.gz` and
`gs://mybucket2/path/to/file2.tar.gz`. `keep_input_artifact` is true, so the
GCE image won't be deleted after the export.

In order for this example to work, the account associated with `account.json`
must have write access to both `gs://mybucket1/path/to/file1.tar.gz` and
`gs://mybucket2/path/to/file2.tar.gz`.

**JSON**

```json
{
  "builders": [
    {
      "type": "googlecompute",
      "account_file": "account.json",
      "project_id": "my-project",
      "source_image": "debian-7-wheezy-v20150127",
      "zone": "us-central1-a"
    }
  ],
  "post-processors": [
    {
      "type": "googlecompute-export",
      "paths": [
        "gs://mybucket1/path/to/file1.tar.gz",
        "gs://mybucket2/path/to/file2.tar.gz"
      ],
      "keep_input_artifact": true
    }
  ]
}
```


**HCL2**

```hcl

  source "googlecompute" "example" {
    account_file = "account.json"
    project_id = "my-project"
    source_image = "debian-7-wheezy-v20150127"
    zone = "us-central1-a"
  }

  build {
    sources = ["source.googlecompute.example"]

    post-processor "googlecompute-export" {
      paths = [
        "gs://mybucket1/path/to/file1.tar.gz",
        "gs://mybucket2/path/to/file2.tar.gz"
      ]
      keep_input_artifact = true
    }
  }
```
