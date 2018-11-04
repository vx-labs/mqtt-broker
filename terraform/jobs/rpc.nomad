job "mqtt-rpc" {
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

    task "rpc" {
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
        args       = ["-r", "9090", "--nomad", "--use-consul"]
        force_pull = true

        port_map {
          health        = 9000
          broker        = 9001
          BrokerService = 9090
        }
      }

      resources {
        cpu    = 200
        memory = 128

        network {
          mbits = 10
          port  "BrokerService"{}
          port  "broker"{}
          port  "health"{}
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
        name = "BrokerService"
        port = "BrokerService"
        tags = ["leader"]

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
