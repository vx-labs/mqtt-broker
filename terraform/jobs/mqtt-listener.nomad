job "mqtt-listener" {
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

  group "tcp-listener" {
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

    task "tcp-listener" {
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
        logging {
          type = "fluentd"

          config {
            fluentd-address = "localhost:24224"
            tag             = "mqtt-listener"
          }
        }

        image      = "quay.io/vxlabs/mqtt-listener:${broker_version}"
        args       = ["-t", "1883"]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          mqtt    = 1883
          service = 4000
          gossip  = 3100
        }
      }

      resources {
        cpu    = 200
        memory = 128

        network {
          mbits = 10
          port  "mqtt"{}
          port  "broker"{}
          port  "cluster"{}
          port  "health"{}
          port  "service"{}
          port  "gossip"{}
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
        name = "tcp-listener"
        port = "mqtt"
        tags = ["urlprefix-:1883 proto=tcp"]

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

  group "tls-listener" {
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

    task "tcp-listener" {
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
        image      = "quay.io/vxlabs/mqtt-listener:${broker_version}"
        args       = ["-s", "8883"]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          service = 4000
          gossip  = 3100
          mqtts   = 8883
        }
      }

      resources {
        cpu    = 200
        memory = 128

        network {
          mbits = 10
          port  "mqtts"{}
          port  "broker"{}
          port  "cluster"{}
          port  "health"{}
          port  "service"{}
          port  "gossip"{}
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
        name = "tls-listener"
        port = "mqtts"
        tags = ["urlprefix-:8883 proto=tcp"]

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

  group "wss-listener" {
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

    task "wss-listener" {
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
        image      = "quay.io/vxlabs/mqtt-listener:${broker_version}"
        args       = ["-w", "8008"]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          service = 4000
          gossip  = 3100
          wss     = 8008
        }
      }

      resources {
        cpu    = 200
        memory = 128

        network {
          mbits = 10
          port  "wss" {}
          port  "broker"{}
          port  "cluster"{}
          port  "health"{}
          port  "service"{}
          port  "gossip"{}
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
