resource "nomad_job" "auth" {
  jobspec = templatefile("${path.module}/consumer.nomad",
    {
      service_name         = "mqtt-auth",
      replica_count        = 3,
      service_image        = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
      args                 = ["service", "auth"],
      exposed_service_name = "",
    },
  )
}
