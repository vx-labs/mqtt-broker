resource "nomad_job" "router" {
  jobspec = templatefile("${path.module}/consumer.nomad",
    {
      service_name         = "mqtt-router",
      replica_count        = 3,
      service_image        = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
      args                 = ["service", "router"],
      exposed_service_name = "",
    },
  )
}
