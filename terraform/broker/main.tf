provider "nomad" {}

variable "broker_version" {}

resource "nomad_job" "broker" {
  jobspec = templatefile("${path.module}/mqtt-broker.nomad", { broker_version = "${var.broker_version}"})
}
