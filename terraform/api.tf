resource "nomad_job" "api" {
  jobspec = templatefile("${path.module}/mqtt-api.nomad", { broker_version = "${var.image_tag}"})
}
