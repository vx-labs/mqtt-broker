resource "nomad_job" "broker" {
  jobspec = templatefile("${path.module}/consumer.nomad",
    {
      service_name         = "mqtt-broker",
      replica_count        = 3,
      service_image        = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
      args                 = ["service", "broker"],
      exposed_service_name = "broker",
    },
  )
}
