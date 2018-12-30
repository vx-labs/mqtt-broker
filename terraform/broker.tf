provider "nomad" {}

variable "broker_version" {}

data "template_file" "tcp-job" {
  template = "${file("${path.module}/jobs/tcp.nomad")}"

  vars {
    broker_version = "${var.broker_version}"
  }
}

data "template_file" "tls-job" {
  template = "${file("${path.module}/jobs/tls.nomad")}"

  vars {
    broker_version = "${var.broker_version}"
  }
}

data "template_file" "wss-job" {
  template = "${file("${path.module}/jobs/wss.nomad")}"

  vars {
    broker_version = "${var.broker_version}"
  }
}

data "template_file" "rpc-job" {
  template = "${file("${path.module}/jobs/rpc.nomad")}"

  vars {
    broker_version = "${var.broker_version}"
  }
}

resource "nomad_job" "mqtt-tcp" {
  jobspec = "${data.template_file.tcp-job.rendered}"
}

resource "nomad_job" "mqtt-tls" {
  jobspec = "${data.template_file.tls-job.rendered}"
}

resource "nomad_job" "mqtt-wss" {
  jobspec = "${data.template_file.wss-job.rendered}"
}

resource "nomad_job" "mqtt-rpc" {
  jobspec = "${data.template_file.rpc-job.rendered}"
}
