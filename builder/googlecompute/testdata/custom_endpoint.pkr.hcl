# Copyright IBM Corp. 2013, 2025
# SPDX-License-Identifier: MPL-2.0

// This template is used for acceptance testing the custom_endpoints and
// universe_domain features. The placeholders will be replaced by the Go test runner.
variable "project_id" {
	type    = string
	default = env("GOOGLE_PROJECT_ID")
}

variable "zone" {
  type    = string
  default = "us-central1-a"
}

variable "compute_custom_endpoints" {
  type    = string
  default = ""
}

variable "universe_domain" {
  type    = string
  default = ""
}

variable "image_name" {
  type    = string
  default = ""
}

source "googlecompute" "acctest" {
	project_id          = var.project_id
	zone                = var.zone
	source_image_family = "centos-stream-9"
	image_name          = var.image_name
	machine_type        = "n2-standard-2"
	ssh_username        = "packer"

	custom_endpoints = {
		"compute": var.compute_custom_endpoints
	}
	universe_domain = var.universe_domain
}

build {
	sources = ["source.googlecompute.acctest"]

	provisioner "shell" {
		inline = ["echo 'Hello, Packer!' > /tmp/test.txt"]
	}
}
