resource "nomad_job" "listener" {
  jobspec = templatefile("${path.module}/mqtt-listener.nomad", { broker_version = "${var.image_tag}"})
}
