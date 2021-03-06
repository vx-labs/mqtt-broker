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
    count = "${replica_count}"
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
%{ if exposed_service_name != "" }
          "--${exposed_service_name}-cluster-bind-port=3100",
          "--${exposed_service_name}-bind-port=4000",
          "--${exposed_service_name}-cluster_rpc-bind-port=3200",
%{ endif }
        ]
        force_pull = true

        port_map {
          health  = 9000
%{ if exposed_service_name != "" }
          rpc            = 4000
          cluster     = 3100
          cluster_rpc = 3200
%{ endif }
        }
      }

      resources {
        cpu    = ${cpu}
        memory = ${memory}

        network {
          mbits = 10
          port  "health"{}
%{ if exposed_service_name != "" }
          port rpc            {}
          port cluster     {}
          port cluster_rpc {}
%{ endif }
        }
      }

      service {
        name = "${exposed_service_name}"
        port = "cluster"
        tags = ["cluster"]
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
        port = "cluster_rpc"
        tags = ["cluster_rpc"]
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
        port = "rpc"
        tags = ["rpc"]

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
