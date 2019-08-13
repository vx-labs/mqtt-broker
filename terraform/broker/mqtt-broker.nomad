job "mqtt-broker" {
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

  group "broker" {
    vault {
      policies      = ["nomad-tls-storer"]
      change_mode   = "signal"
      change_signal = "SIGUSR1"
      env           = false
    }

    count = 3

    restart {
      attempts = 10
      interval = "5m"
      delay    = "15s"
      mode     = "delay"
    }

    ephemeral_disk {
      size = 300
    }

    task "broker" {
      driver = "docker"

      env {
        CONSUL_HTTP_ADDR          = "172.17.0.1:8500"
        AUTH_HOST                 = "172.17.0.1:4141"
        JAEGER_SAMPLER_TYPE       = "const"
        JAEGER_SAMPLER_PARAM      = "1"
        JAEGER_REPORTER_LOG_SPANS = "true"
        JAEGER_AGENT_HOST         = "$${NOMAD_IP_health}"
      }
      template {
        destination = "local/proxy.conf"
        env = true
        data = <<EOH
{{ with secret "secret/data/vx/mqtt" }}
http_proxy="{{ .Data.http_proxy }}"
https_proxy="{{ .Data.http_proxy }}"
JWT_SIGN_KEY="{{ .Data.jwt_sign_key }}"
no_proxy="10.0.0.0/8,172.16.0.0/12"
{{ end }}
EOH
      }
      config {
        logging {
          type = "fluentd"

          config {
            fluentd-address = "localhost:24224"
            tag             = "mqtt-broker"
          }
        }

        image      = "quay.io/vxlabs/mqtt-broker:${broker_version}"
        args       = ["--cluster-bind-port=3500", "--gossip-bind-port=3100", "--service-bind-port=4000"]
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
    }
  }
}
