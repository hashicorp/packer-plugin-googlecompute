# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "Google Cloud Platform"
  description = "The googlecompute plugin can be used with HashiCorp Packer to create custom images on GCE."
  identifier = "packer/hashicorp/googlecompute"
  flags = ["hcp-ready"]
  component {
    type = "builder"
    name = "Google Cloud Platform"
    slug = "googlecompute"
  }
  component {
    type = "post-processor"
    name = "Google Cloud Platform Image Import"
    slug = "googlecompute-import"
  }
  component {
    type = "post-processor"
    name = "Google Cloud Platform Image Exporter"
    slug = "googlecompute-export"
  }
}
