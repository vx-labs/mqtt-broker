resource "nomad_job" "topics" {
  jobspec = templatefile("${path.module}/consumer.nomad",
    {
      service_name         = "mqtt-topics",
      replica_count        = 3,
      service_image        = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
      args                 = ["service", "topics"],
      exposed_service_name = "topics",
    },
  )
}
