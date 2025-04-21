variable "project_id" {
  type = string
}

variable "secret_name" {
  type = string
  default = "packer-test-secret"
}

variable "key" {
  type = string
  default = "foo"
}

data "googlecompute-secretsmanager" "test" {
  project_id  = var.project_id
  name        = var.secret_name
  key         = var.key
}

locals {
  value = data.googlecompute-secretsmanager.test.value
  payload = data.googlecompute-secretsmanager.test.payload
}

source "null" "basic-example" {
  communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  provisioner "shell-local" {
    inline = [
      "echo secret value: ${local.value}",
      "echo secret payload: ${local.payload}",
    ]
  }
}
