job "${service_name}" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "30s"
    healthy_deadline = "10m"
    progress_deadline = "30m"
    health_check     = "task_states"
    auto_revert      = true
    canary           = 0
  }

  group "${service_name}" {
     vault {
      policies      = ["nomad-tls-storer"]
      change_mode   = "signal"
      change_signal = "SIGUSR1"
      env           = false
    }
    count = ${replica_count}
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

    task "${service_name}" {
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
PSK_PASSWORD="{{ .Data.static_tokens }}"
TLS_CERTIFICATE="{{ env "NOMAD_TASK_DIR" }}/cert.pem"
TLS_PRIVATE_KEY="{{ env "NOMAD_TASK_DIR" }}/key.pem"
TLS_CA_CERTIFICATE="{{ env "NOMAD_TASK_DIR" }}/ca.pem"
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
{{ with secret $path $cn $ipsans "ttl=480h" }}{{ .Data.certificate }}{{ end }}
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
{{ with secret $path $cn $ipsans "ttl=480h" }}{{ .Data.private_key }}{{ end }}
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
{{ with secret $path $cn $ipsans "ttl=480h" }}{{ .Data.issuing_ca }}{{ end }}
EOH
      }
      config {
        logging {
          type = "fluentd"

          config {
            fluentd-address = "localhost:24224"
            tag             = "${service_name}"
          }
        }

        image      = "${service_image}"
        args       = [
%{ for arg in args }
          "${arg}",
%{ endfor }
          "--cluster-bind-port=3500",
%{ if exposed_service_name != "" }
          "--${exposed_service_name}gossip-bind-port=3100",
          "--${exposed_service_name}-bind-port=4000",
          "--${exposed_service_name}gossiprpc-bind-port=3200",
%{ endif }
        ]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
%{ if exposed_service_name != "" }
          ${exposed_service_name}            = 4000
          ${exposed_service_name}gossip     = 3100
          ${exposed_service_name}gossiprpc = 3200
%{ endif }
        }
      }

      resources {
        cpu    = ${cpu}
        memory = ${memory}

        network {
          mbits = 10
          port  "cluster"{}
          port  "health"{}
%{ if exposed_service_name != "" }
          port ${exposed_service_name}            {}
          port ${exposed_service_name}gossip     {}
          port ${exposed_service_name}gossiprpc {}
%{ endif }
        }
      }

      service {
        name = "${exposed_service_name}gossip"
        port = "${exposed_service_name}gossip"

        check {
          type     = "http"
          path     = "/health"
          port     = "health"
          interval = "5s"
          timeout  = "2s"
        }
      }
      service {
        name = "${exposed_service_name}gossiprpc"
        port = "${exposed_service_name}gossiprpc"

        check {
          type     = "http"
          path     = "/health"
          port     = "health"
          interval = "5s"
          timeout  = "2s"
        }
      }
      service {
        name = "${exposed_service_name}"
        port = "${exposed_service_name}"

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
