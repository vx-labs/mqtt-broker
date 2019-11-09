provider "nomad" {}

variable "router_version" {}

resource "nomad_job" "router" {
  jobspec = templatefile("${path.module}/mqtt-router.nomad", { broker_version = "${var.router_version}"})
}
