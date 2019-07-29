job "mqtt-api" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "30s"
    healthy_deadline = "3m"
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

    count = 1

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
        TLS_CN                    = "broker-api.iot.cloud.vx-labs.net"
        CONSUL_HTTP_ADDR          = "172.17.0.1:8500"
        AUTH_HOST                 = "172.17.0.1:4141"
        JAEGER_SAMPLER_TYPE       = "const"
        JAEGER_SAMPLER_PARAM      = "1"
        JAEGER_REPORTER_LOG_SPANS = "true"
        JAEGER_AGENT_HOST         = "$${NOMAD_IP_health}"
      }

      config {
        logging {
          type = "fluentd"

          config {
            fluentd-address = "localhost:24224"
            tag             = "mqtt-api"
          }
        }

        image      = "quay.io/vxlabs/mqtt-api:${broker_version}"
        args       = ["--tls-port", "3000", "--cluster-bind-port=3500", "--gossip-bind-port=3100", "--service-bind-port=4000"]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          service = 4000
          gossip  = 3100
          https   = 3000
        }
      }

      resources {
        cpu    = 200
        memory = 64

        network {
          mbits = 10
          port  "mqtts"{}
          port  "broker"{}
          port  "cluster"{}
          port  "health"{}
          port  "service"{}
          port  "gossip"{}
          port  "https"{}
        }
      }

      service {
        name = "cluster"
        port = "cluster"

        check {
          type     = "http"
          path     = "/health"
          port     = "health"
          interval = "5s"
          timeout  = "2s"
        }
      }

      service {
        name = "tls-api"
        port = "https"
        tags = ["urlprefix-broker-api.iot.cloud.vx-labs.net/ proto=tcp+sni"]

        check {
          type     = "http"
          path     = "/health"
          port     = "health"
          interval = "5s"
          timeout  = "2s"
        }
      }

      service {
        name = "mqtt-metrics"
        port = "health"

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
