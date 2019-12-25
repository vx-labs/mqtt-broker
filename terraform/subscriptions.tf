resource "nomad_job" "subscriptions" {
  jobspec = templatefile("${path.module}/consumer.nomad",
    {
      service_name         = "mqtt-subscriptions",
      replica_count        = 3,
      service_image        = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
      args                 = ["service", "subscriptions"],
      exposed_service_name = "subscriptions",
    },
  )
}
