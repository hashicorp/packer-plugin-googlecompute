The Parameter Manager data source provides the capability to retrieve the parameters stored in google cloud along with revealed secrets

-> **Note:** Data sources is a feature exclusively available to HCL2 templates.

Basic examples of usage:

```hcl
data "googlecompute-parametermanager" "basic-example" {
  project_id = "debian-cloud"
  name       = "packer_test_parameter"
  key        = "packer_test_key"
  version    = "1"
  location   = "us-east1"    
}

# usage example of the data source output
locals {
  value = data.googlecompute-parametermanager.basic-example.value
  payload = data.googlecompute-parametermanager.basic-example.payload
}
```

Reading key-value pairs from JSON back into a native Packer map can be accomplished
with the [jsondecode() function](/packer/docs/templates/hcl_templates/functions/encoding/jsondecode).

Reading key-value pairs from YAML back into a native Packer map can be accomplished
with the [yamldecode() function](/packer/docs/templates/hcl_templates/functions/encoding/yamldecode).

## Configuration Reference

### Required

<!-- Code generated from the comments of the Config struct in datasource/parametermanager/data.go; DO NOT EDIT MANUALLY -->

- `project_id` (string) - The Google Cloud project ID where the parameter is stored.

- `name` (string) - The name of the parameter within the Parameter Manager.

- `version` (string) - The version of the parameter within the Parameter Manager.

<!-- End of code generated from the comments of the Config struct in datasource/parametermanager/data.go; -->


### Optional

<!-- Code generated from the comments of the Config struct in datasource/parametermanager/data.go; DO NOT EDIT MANUALLY -->

- `location` (string) - The location in which parameter is stored. Defaults to "global" if not specified.

- `key` (string) - A specific key to extract from the parameter payload if it's a JSON or YAML object.

<!-- End of code generated from the comments of the Config struct in datasource/parametermanager/data.go; -->


## Output Data

<!-- Code generated from the comments of the DatasourceOutput struct in datasource/parametermanager/data.go; DO NOT EDIT MANUALLY -->

- `payload` (string) - The raw string payload of the parameter version.

- `value` (string) - The value extracted using the 'key', if provided.

<!-- End of code generated from the comments of the DatasourceOutput struct in datasource/parametermanager/data.go; -->


## Authentication

To authenticate with GCE, this data-source supports everything the plugin does.
To get more information on this, refer to the plugin's description page, under
the [authentication](/packer/integrations/hashicorp/googlecompute#authentication) section.
