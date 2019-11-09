provider "nomad" {}

variable "messages_version" {}

resource "nomad_job" "messages" {
  jobspec = templatefile("${path.module}/mqtt-messages.nomad", { broker_version = "${var.messages_version}"})
}
