module "kv" {
  source               = "./modules/task"
  service_name         = "mqtt-kv"
  replica_count        = 3
  image_tag            = var.image_tag
  args                 = ["service", "kv"]
  exposed_service_name = "kv"
  memory               = "256"
}
