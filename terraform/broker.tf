provider "nomad" {}

variable "broker_version" {}

resource "nomad_job" "mqtt-tcp" {
  jobspec = templatefile("${path.module}/jobs/tcp.nomad", { broker_version = "${var.broker_version}"})
}

resource "nomad_job" "mqtt-tls" {
  jobspec = templatefile("${path.module}/jobs/tls.nomad", { broker_version = "${var.broker_version}"})
}

resource "nomad_job" "mqtt-wss" {
  jobspec = templatefile("${path.module}/jobs/wss.nomad", { broker_version = "${var.broker_version}"})
}
