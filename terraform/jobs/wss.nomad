job "mqtt-wss" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "3m"
    health_check     = "checks"
    auto_revert      = true
    canary           = 0
  }

  group "broker" {
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

    task "wss" {
      driver = "docker"

      env {
        CONSUL_HTTP_ADDR          = "172.17.0.1:8500"
        AUTH_HOST                 = "172.17.0.1:4141"
        JAEGER_SAMPLER_TYPE       = "const"
        JAEGER_SAMPLER_PARAM      = "1"
        JAEGER_REPORTER_LOG_SPANS = "true"
        JAEGER_AGENT_HOST         = "$${NOMAD_IP_health}"
      }

      config {
        image      = "quay.io/vxlabs/mqtt-broker:${broker_version}"
        args       = ["-w", "8008", "-r", "9091", "--use-vault", "--use-vx-auth", "--nomad", "--use-consul"]
        force_pull = true

        port_map {
          health = 9000
          broker = 9090
          rpc    = 9091
          wss    = 8008
        }
      }

      resources {
        cpu    = 100
        memory = 128

        network {
          mbits = 10
          port  "wss" {}
          port  "broker"{}
          port  "rpc" {}
          port  "health"{}
        }
      }

      service {
        name = "rpc"
        port = "rpc"

        check {
          type     = "http"
          path     = "/health"
          port     = "health"
          interval = "5s"
          timeout  = "2s"
        }
      }

      service {
        name = "broker"
        port = "broker"

        check {
          type     = "http"
          path     = "/health"
          port     = "health"
          interval = "5s"
          timeout  = "2s"
        }
      }

      service {
        name = "wss-listener"
        port = "wss"
        tags = ["urlprefix-broker.iot.cloud.vx-labs.net/ proto=tcp+sni"]

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
