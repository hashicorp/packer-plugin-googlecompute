# Copyright IBM Corp. 2013, 2025
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    googlecompute = {
      version = "~> v1.0"
      source  = "github.com/hashicorp/googlecompute"
    }
  }
}

variable "project_id" {
  type = string
}

variable "family" {
  type = string
}

data "googlecompute-image" "example" {
  project_id = var.project_id
  filters = "family=debian-12 AND labels.public-image=true"
  most_recent = true
}

source "null" "ex" {
  communicator = "none"
}

build {
  sources = ["source.null.ex"]
  provisioner "shell-local" {
    inline = [
      "echo ${data.googlecompute-image.example.id}",
      "echo ${data.googlecompute-image.example.name}",
    ]
  }
}