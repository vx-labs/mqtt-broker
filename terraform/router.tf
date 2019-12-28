module "router" {
  source               = "./modules/task"
  service_name         = "mqtt-router"
  image_tag            = var.image_tag
  args                 = ["service", "router"]
  exposed_service_name = "router"
}
