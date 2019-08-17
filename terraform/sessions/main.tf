provider "nomad" {}

variable "sessions_version" {}

resource "nomad_job" "listener" {
  jobspec = templatefile("${path.module}/mqtt-sessions.nomad", { broker_version = "${var.sessions_version}"})
}
