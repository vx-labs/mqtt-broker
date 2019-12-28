provider "nomad" {
  version = "~> 1.4"
}

variable service_name {}
variable memory {
  default = "32"
}
variable replica_count {
  default = "2"
}
variable image_repository {
  default = "quay.io/vxlabs/mqtt-broker"
}
variable image_tag {}
variable args {
  default = []
}
variable exposed_service_name {
  default = ""
}

resource "nomad_job" "messages" {
  jobspec = templatefile("${path.module}/template.nomad",
    {
      service_name         = var.service_name,
      replica_count        = var.replica_count,
      service_image        = "${var.image_repository}:${var.image_tag}",
      args                 = var.args,
      exposed_service_name = var.exposed_service_name,
      memory               = var.memory,
    },
  )
}
