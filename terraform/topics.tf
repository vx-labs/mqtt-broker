module "topics" {
  source               = "./modules/task"
  service_name         = "mqtt-topics"
  image_tag            = var.image_tag
  args                 = ["service", "topics"]
  exposed_service_name = "topics"
}
