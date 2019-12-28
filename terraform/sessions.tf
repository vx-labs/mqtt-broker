module "sessions" {
  source               = "./modules/task"
  service_name         = "mqtt-sessions"
  image_tag            = var.image_tag
  args                 = ["service", "sessions"]
  exposed_service_name = "sessions"
}
