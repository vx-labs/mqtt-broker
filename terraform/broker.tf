provider "nomad" {}

variable "broker_version" {}
variable "listener_version" {}

resource "nomad_job" "iot" {
  jobspec = templatefile("${path.module}/jobs/mqtt-broker.nomad", { broker_version = "${var.broker_version}"})
}

resource "nomad_job" "listener" {
  jobspec = templatefile("${path.module}/jobs/mqtt-listener.nomad", { broker_version = "${var.listener_version}"})
}

