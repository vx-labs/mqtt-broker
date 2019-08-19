provider "nomad" {}

variable "subscriptions_version" {}

resource "nomad_job" "listener" {
  jobspec = templatefile("${path.module}/mqtt-subscriptions.nomad", { broker_version = "${var.subscriptions_version}"})
}
