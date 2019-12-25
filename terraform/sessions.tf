resource "nomad_job" "sessions" {
  jobspec = templatefile("${path.module}/consumer.nomad",
    {
      service_name = "mqtt-sessions",
      replica_count = 3,
      service_image = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
      exposed_service_name = "sessions",
      args = ["service", "sessions"],
    },
  )
}
