# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "project" {
  type    = string
  default = env("GOOGLE_PROJECT_ID")
}

variable "service_account_file" {
  type    = string
  default = env("GOOGLE_APPLICATION_CREDENTIALS")
}

variable "ssh_username" {
  type    = string
  default = "packer"
}

variable "zone" {
  type    = string
  default = "us-central1-a"
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "googlecompute" "test" {
  account_file        = var.service_account_file
  image_name          = "packer-multiple-disks-test-${local.timestamp}"
  project_id          = var.project
  source_image_family = "centos-stream-9"
  ssh_username        = var.ssh_username
  skip_create_image   = true
  machine_type        = "n2-standard-2"
  zone                = var.zone

  disk_attachment {
    attachment_mode = "READ_WRITE"
    volume_type     = "pd-standard"
    volume_size     = 25
    interface_type  = "SCSI"
  }

  disk_attachment {
    volume_type = "scratch"
    volume_size = 375
  }
}

build {
  sources = ["source.googlecompute.test"]

  provisioner "shell" {
    inline = [
      "set -ex",
      "ls -la /dev/disk/by-id/google-local-ssd-0",
      "ls -la /dev/disk/by-id/google-persistent-disk-2"
    ]
  }
}

