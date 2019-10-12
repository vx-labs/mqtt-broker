provider "nomad" {}

variable "queues_version" {}

resource "nomad_job" "listener" {
  jobspec = templatefile("${path.module}/mqtt-queues.nomad", { broker_version = "${var.queues_version}"})
}
