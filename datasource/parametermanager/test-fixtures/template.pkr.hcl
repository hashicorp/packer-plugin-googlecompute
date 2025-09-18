variable "project_id" {
  type = string
}

variable "parameter_name" {
  type = string
  default = "packer-test-parameter"
}

variable "key" {
  type = string
  default = "foo"
}

variable "format" {
  type = string
}

variable "parse_payload" {
  type    = bool
  default = true
}

variable "location" {
  type    = string
  default = "global"
}

data "googlecompute-parametermanager" "test" {
  project_id  = var.project_id
  name       = var.parameter_name
  version    = "1"
  key        = var.key
  location = var.location
}


# usage example of the data source output
locals {
    payload = var.format == "JSON" ? jsondecode(data.googlecompute-parametermanager.test.payload) : (
    var.format == "YAML" ? yamldecode(data.googlecompute-parametermanager.test.payload) : { "key" : data.googlecompute-parametermanager.test.payload}
  )

  unformatted_payload = data.googlecompute-parametermanager.test.payload
  project_id = var.parse_payload ? local.payload.project_id : ""
  test_value   = data.googlecompute-parametermanager.test.value
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
      "echo Unformatted Payload is '${local.unformatted_payload}'",
      "echo Parameter value using key is ${local.test_value}",
      "echo Project id parsed from ${var.format} is ${local.project_id}"
    ]
  }
}