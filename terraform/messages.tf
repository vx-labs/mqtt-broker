module "messages" {
  source        = "./modules/task"
  service_name  = "mqtt-messages"
  image_tag     = var.image_tag
  replica_count = 3
  args = [
    "service", "messages",
    "--initial-stream-config", "messages:3",
    "--initial-stream-config", "events:3",
  ]
  exposed_service_name = "messages"
  memory               = "128"
}
