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
  image_name          = "packer-persistent-disks-region-test-${local.timestamp}"
  project_id          = var.project
  source_image_family = "centos-stream-9"
  ssh_username        = var.ssh_username
  skip_create_image   = true
  zone                = var.zone
  disk_attachment {
    attachment_mode = "READ_WRITE"
    volume_type     = "pd-standard"
    volume_size     = 200
    interface_type  = "SCSI"
    replica_zones   = ["us-central1-b"]
  }
}

build {
  sources = ["source.googlecompute.test"]

  provisioner "shell" {
    # persistent-disk-0 is already reserved for the boot disk, the ones we add will start at 1
    inline = ["ls -la /dev/disk/by-id/google-persistent-disk-1"]
  }
}
