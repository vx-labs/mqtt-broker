provider "nomad" {}

variable "api_version" {}

resource "nomad_job" "api" {
  jobspec = templatefile("${path.module}/mqtt-api.nomad", { broker_version = "${var.api_version}"})
}
