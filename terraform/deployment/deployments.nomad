job "deployments" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "90s"
    healthy_deadline = "120s"
    health_check     = "checks"
    auto_revert      = true
    canary           = 0
  }

  group "tls-api" {
    vault {
      policies      = ["nomad-tls-storer"]
      change_mode   = "signal"
      change_signal = "SIGUSR1"
      env           = false
    }

    count = 2

    restart {
      attempts = 10
      interval = "5m"
      delay    = "15s"
      mode     = "delay"
    }

    ephemeral_disk {
      size = 300
    }

    task "server" {
      driver = "docker"

      env {
        TLS_CN           = "deployments.iot.cloud.vx-labs.net"
        CONSUL_HTTP_ADDR = "${NOMAD_IP_service}:8500"
        VAULT_ADDR       = "http://active.vault.service.consul:8200/"
        NOMAD_ADDR       = "http://servers.nomad.discovery.par1.vx-labs.net:4646"
      }

      template {
        destination = "local/proxy.conf"
        env = true
        data = <<EOH
{{with secret "secret/data/vx/mqtt"}}
http_proxy="{{.Data.http_proxy}}"
https_proxy="{{.Data.http_proxy}}"
LE_EMAIL="{{.Data.acme_email}}"
no_proxy="10.0.0.0/8,172.16.0.0/12,*.service.consul"
{{end}}
        EOH
      }

      config {
        image      = "quay.io/jbonachera/nomad-quay-deployer:latest"
        force_pull = true

        port_map {
          service  = 8081
          health  = 9000
        }
      }

      resources {
        cpu    = 200
        memory = 32

        network {
          mbits = 10
          port  "service"{}
          port  "health"{}
        }
      }

      service {
        name = "tls-deployments"
        port = "service"
        tags = ["urlprefix-deployments.iot.cloud.vx-labs.net/ proto=tcp+sni"]

        check {
          type     = "http"
          path     = "/health"
          port     = "health"
          interval = "5s"
          timeout  = "2s"
        }
      }
    }
  }
}
