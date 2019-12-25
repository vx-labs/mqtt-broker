resource "nomad_job" "messages" {
  jobspec = templatefile("${path.module}/consumer.nomad",
    {
      service_name  = "mqtt-messages",
      replica_count = 3,
      service_image = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
      args = ["service", "messages",
        "--initial-stream-config", "messages:3",
        "--initial-stream-config", "events:3",
      ],
      exposed_service_name = "messages",
    },
  )
}
