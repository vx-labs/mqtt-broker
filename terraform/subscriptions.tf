module "subscriptions" {
  source               = "./modules/task"
  service_name         = "mqtt-subscriptions"
  image_tag            = var.image_tag
  args                 = ["service", "subscriptions"]
  exposed_service_name = "subscriptions"
}
