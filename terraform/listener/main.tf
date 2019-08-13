provider "nomad" {}

variable "listener_version" {}

resource "nomad_job" "listener" {
  jobspec = templatefile("${path.module}/mqtt-listener.nomad", { broker_version = "${var.listener_version}"})
}
