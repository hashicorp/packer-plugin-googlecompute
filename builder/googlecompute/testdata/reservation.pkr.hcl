variable "image_name" {
  type = string
}

variable "reservation_name" {
  type = string
}

variable "project_id" {
  type = string
}

source "googlecompute" "reservation-test" {
  project_id               = var.project_id
  source_image_family      = "ubuntu-2204-lts"
  source_image_project_id  = ["ubuntu-os-cloud"]
  zone                     = "us-central1-a"
  tags                     = ["packer-test"]
  network                  = "default"
  ssh_username             = "packer"
  image_name          = var.image_name
  machine_type        = "n1-standard-1"

  reservation_affinity {
    consume_reservation_type = "SPECIFIC_RESERVATION"
    key                      = "compute.googleapis.com/reservation-name"
    values                   = [var.reservation_name]
  }
}

build {
  sources = ["source.googlecompute.reservation-test"]
}
