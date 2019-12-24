resource "nomad_job" "kv" {
  jobspec = templatefile("${path.module}/consumer.nomad", {
    service_name         = "mqtt-kv",
    replica_count        = 3,
    service_image        = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
    args                 = ["service", "kv"],
    exposed_service_name = "kv",
    },
  )
}
