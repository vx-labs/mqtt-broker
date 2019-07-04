provider "nomad" {}

variable "broker_version" {}

resource "nomad_job" "iot" {
  jobspec = templatefile("${path.module}/jobs/iot.nomad", { broker_version = "${var.broker_version}"})
}

