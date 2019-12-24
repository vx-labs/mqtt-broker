job "mqtt-kv" {
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

  group "kv" {
     vault {
      policies      = ["nomad-tls-storer"]
      change_mode   = "signal"
      change_signal = "SIGUSR1"
      env           = false
    }
    count = 3
    constraint {
        distinct_hosts = true
    }

    restart {
      attempts = 10
      interval = "5m"
      delay    = "15s"
      mode     = "delay"
    }

    ephemeral_disk {
      size = 300
    }

    task "store" {
      driver = "docker"
      env {
        CONSUL_HTTP_ADDR = "$${NOMAD_IP_health}:8500"
        VAULT_ADDR       = "http://active.vault.service.consul:8200/"
      }

      template {
        destination = "local/proxy.conf"
        env         = true

        data = <<EOH
{{with secret "secret/data/vx/mqtt"}}
http_proxy="{{.Data.http_proxy}}"
https_proxy="{{.Data.http_proxy}}"
LE_EMAIL="{{.Data.acme_email}}"
JWT_SIGN_KEY="{{ .Data.jwt_sign_key }}"
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
            tag             = "mqtt-kv"
          }
        }

        image      = "quay.io/vxlabs/mqtt-kv:${broker_version}"
        args       = [
          "--cluster-bind-port=3500",
          "--kv_gossip-bind-port=3100",
          "--kv-bind-port=4000",
          "--kv_gossip_rpc-bind-port=3200",
        ]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          kv = 4000
          kv_gossip  = 3100
          kv_gossip_rpc  = 3200
        }
      }

      resources {
        cpu    = 200
        memory = 128

        network {
          mbits = 10
          port  "cluster"{}
          port  "health"{}
          port  "kv"{}
          port  "kv_gossip"{}
          port  "kv_gossip_rpc"{}
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
    }
  }
}
