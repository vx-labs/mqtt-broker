provider "nomad" {}

variable "topics_version" {}

resource "nomad_job" "listener" {
  jobspec = templatefile("${path.module}/mqtt-topics.nomad", { broker_version = "${var.topics_version}"})
}
