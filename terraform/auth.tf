module "auth" {
  source               = "./modules/task"
  service_name         = "mqtt-auth"
  image_tag            = var.image_tag
  args                 = ["service", "auth"]
  exposed_service_name = "auth"
}
