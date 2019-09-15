provider "nomad" {}

resource "nomad_job" "deployments" {
  jobspec = "${file("${path.module}/deployments.nomad")}"
}
