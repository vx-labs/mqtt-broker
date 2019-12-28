module "broker" {
  source               = "./modules/task"
  service_name         = "mqtt-broker"
  image_tag            = var.image_tag
  args                 = ["service", "broker"]
  exposed_service_name = "broker"
}
