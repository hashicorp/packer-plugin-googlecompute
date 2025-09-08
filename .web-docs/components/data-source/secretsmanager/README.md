The Secrets Manager data source provides information about a Secrets Manager secret version,
including its value and metadata.

-> **Note:** Data sources is a feature exclusively available to HCL2 templates.

Basic examples of usage:

```hcl
data "googlecompute-secretsmanager" "basic-example" {
  project_id = "debian-cloud"
  name       = "packer_test_secret"
  key        = "packer_test_key"
}

# usage example of the data source output
locals {
  value = data.googlecompute-secretsmanager.basic-example.value
  payload = data.googlecompute-secretsmanager.basic-example.payload
}
```

Reading key-value pairs from JSON back into a native Packer map can be accomplished
with the [jsondecode() function](/packer/docs/templates/hcl_templates/functions/encoding/jsondecode).

## Configuration Reference

### Required

<!-- Code generated from the comments of the Config struct in datasource/secretsmanager/data.go; DO NOT EDIT MANUALLY -->

- `project_id` (string) - The Google Cloud project ID where the secret is stored.

- `name` (string) - The name of the secret in the secret manager.

<!-- End of code generated from the comments of the Config struct in datasource/secretsmanager/data.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in datasource/secretsmanager/data.go; DO NOT EDIT MANUALLY -->

- `key` (string) - The key to extract from the secret payload.
  If not provided, the entire payload will be returned.

- `version` (string) - The version of the secret to access. Defaults to "latest" if not specified.

- `universe_domain` (string) - Specify the GCP universe to deploy in. The default is "googleapis.com".

- `custom_endpoints` (map[string]string) - Custom service endpoints, typically used to configure the Google provider to
  communicate with GCP-like APIs such as the Cloud Functions emulator.
   Supported keys are `secretmanager`.
  
  Example:
    custom_endpoints = {
      secretmanager = "https://{your-endpoint}/"
    }

<!-- End of code generated from the comments of the Config struct in datasource/secretsmanager/data.go; -->


## Output Data

<!-- Code generated from the comments of the DatasourceOutput struct in datasource/secretsmanager/data.go; DO NOT EDIT MANUALLY -->

- `payload` (string) - The raw string payload of the secret version.

- `value` (string) - The value extracted using the 'key', if provided.

- `checksum` (int64) - The crc32c checksum for the payload.

<!-- End of code generated from the comments of the DatasourceOutput struct in datasource/secretsmanager/data.go; -->


## Authentication

To authenticate with GCE, this data-source supports everything the plugin does.
To get more information on this, refer to the plugin's description page, under
the [authentication](/packer/integrations/hashicorp/googlecompute#authentication) section.
