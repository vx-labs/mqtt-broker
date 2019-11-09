provider "nomad" {}

variable "kv_version" {}

resource "nomad_job" "kv" {
  jobspec = templatefile("${path.module}/mqtt-kv.nomad", { broker_version = "${var.kv_version}"})
}
