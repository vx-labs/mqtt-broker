module "queues" {
  source               = "./modules/task"
  service_name         = "mqtt-queues"
  image_tag            = var.image_tag
  args                 = ["service", "queues"]
  replica_count        = 3
  exposed_service_name = "queues"
  memory               = "128"
}
