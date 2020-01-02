module "endpoints" {
  source               = "./modules/task"
  service_name         = "mqtt-endpoints"
  image_tag            = var.image_tag
  args                 = ["service", "endpoints"]
  exposed_service_name = "endpoints"
}
