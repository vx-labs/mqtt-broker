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
        TLS_CN           = "broker-api.iot.cloud.vx-labs.net"
        CONSUL_HTTP_ADDR = "$${NOMAD_IP_health}:8500"
        VAULT_ADDR       = "http://active.vault.service.consul:8200/"
      }

      template {
        destination = "local/proxy.conf"
        env = true
        data = <<EOH
{{with secret "secret/data/vx/mqtt"}}
http_proxy="{{.Data.http_proxy}}"
https_proxy="{{.Data.http_proxy}}"
LE_EMAIL="{{.Data.acme_email}}"
TLS_CERTIFICATE="{{ env "NOMAD_TASK_DIR" }}//cert.pem"
TLS_PRIVATE_KEY="{{ env "NOMAD_TASK_DIR" }}//key.pem"
TLS_CA_CERTIFICATE="{{ env "NOMAD_TASK_DIR" }}//ca.pem"
no_proxy="10.0.0.0/8,172.16.0.0/12,*.service.consul"
{{end}}
EOH
      }
      template {
        change_mode   = "restart"
        destination = "local/cert.pem"
        splay = "1h"
        data = <<EOH
{{- $cn := printf "common_name=%s" (env "NOMAD_ALLOC_ID") -}}
{{- $ipsans := printf "ip_sans=%s" (env "NOMAD_IP_health") -}}
{{- $path := printf "pki/issue/grpc" -}}
{{ with secret $path $cn $ipsans "ttl=48h" }}{{ .Data.certificate }}{{ end }}
EOH
      }
      template {
        change_mode   = "restart"
        destination = "local/key.pem"
        splay = "1h"
        data = <<EOH
{{- $cn := printf "common_name=%s" (env "NOMAD_ALLOC_ID") -}}
{{- $ipsans := printf "ip_sans=%s" (env "NOMAD_IP_health") -}}
{{- $path := printf "pki/issue/grpc" -}}
{{ with secret $path $cn $ipsans "ttl=48h" }}{{ .Data.private_key }}{{ end }}
EOH
      }
      template {
        change_mode   = "restart"
        destination = "local/ca.pem"
        splay = "1h"
        data = <<EOH
{{- $cn := printf "common_name=%s" (env "NOMAD_ALLOC_ID") -}}
{{- $ipsans := printf "ip_sans=%s" (env "NOMAD_IP_health") -}}
{{- $path := printf "pki/issue/grpc" -}}
{{ with secret $path $cn $ipsans "ttl=48h" }}{{ .Data.issuing_ca }}{{ end }}
EOH
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
        args       = [
          "--tls-port=3000",
          "--cluster-bind-port=3500",
          "--api_gossip-bind-port=3100",
          "--api-bind-port=4000",
        ]

        port_map {
          health  = 9000
          cluster = 3500
          api = 4000
          api_gossip  = 3100
          https   = 3000
        }
      }

      resources {
        cpu    = 200
        memory = 64

        network {
          mbits = 10
          port  "cluster"{}
          port  "health"{}
          port  "api"{}
          port  "api_gossip"{}
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
        tags = [
          "traefik.enable=true",
          "traefik.tcp.routers.api.rule=HostSNI(`broker-api.iot.cloud.vx-labs.net`)",
          "traefik.tcp.routers.api.entrypoints=https",
          "traefik.tcp.routers.api.tls",
          "traefik.tcp.routers.api.tls.passthrough=true"
        ]

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
