Type: `googlecompute-image`

The Google Compute Image data source filters and fetches a GCE image and outputs relevant image metadata for
use with [Google Compute builders](/packer/integrations/hashicorp/googlecompute).

-> **Note:** Data sources is a feature exclusively available to HCL2 templates.

Basic example of usage:

```hcl
data "googlecompute-image" "basic-example" {
  project_id = "debian-cloud"
  filters = "family=debian-12 AND labels.public-image=true"
  most_recent = true
}
```

This configuration selects the most recent GCE image from the `debian-cloud` project that belongs to the`debian-12` family and has the `public-image` label set to `true`.
The data source will fail unless exactly one image is matched. Setting `most_recent = true` ensures only the newest image is selected when multiple matches exist.

## Configuration Reference

<!-- Code generated from the comments of the Config struct in datasource/image/data.go; DO NOT EDIT MANUALLY -->

- `project_id` (string) - The Google Cloud project ID to search for images.

- `filters` (string) - The filter expression to narrow down the image search.
  For example: "name=ubuntu" or "family=ubuntu-2004".
  The exrpressions can be combined with AND/OR like this:
  "name=ubuntu AND family=ubuntu-2004".
  See https://cloud.google.com/sdk/gcloud/reference/topic/filters

- `most_recent` (bool) - If true, the most recent image will be returned.
  If false, an error will be returned if more than one image matches the filters.

- `universe_domain` (string) - Specify the GCP universe to deploy in. The default is "googleapis.com".

- `custom_endpoints` (map[string]string) - Custom service endpoints, typically used to configure the Google provider to
  communicate with GCP-like APIs such as the Cloud Functions emulator.
   Supported keys are `compute`.
  
  Example:
    custom_endpoints = {
      compute = "https://{your-endpoint}/"
    }

<!-- End of code generated from the comments of the Config struct in datasource/image/data.go; -->


## Output Data

<!-- Code generated from the comments of the DatasourceOutput struct in datasource/image/data.go; DO NOT EDIT MANUALLY -->

- `id` (string) - ID

- `name` (string) - Name

- `creation_date` (string) - Creation Date

- `labels` (map[string]string) - Labels

<!-- End of code generated from the comments of the DatasourceOutput struct in datasource/image/data.go; -->


## Authentication

To authenticate with GCE, this data-source supports everything the plugin does.
To get more information on this, refer to the plugin's description page, under
the [authentication](/packer/integrations/hashicorp/googlecompute#authentication) section.
