resource "nomad_job" "queues" {
  jobspec = templatefile("${path.module}/consumer.nomad", {
    service_name         = "mqtt-queues",
    replica_count        = 3,
    service_image        = "quay.io/vxlabs/mqtt-broker:${var.image_tag}",
    args                 = ["service", "queues"],
    exposed_service_name = "queues",
    },
  )
}
