data "google-compute-secret-manager" "test" {
  project = "my-secret"
  name = "my-project"
  version = "1"
}

locals {
  value = data.google-compute-secret-manage.test.value
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
    ]
  }
}
