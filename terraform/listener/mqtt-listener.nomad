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

    task "tcp-listener" {
      driver = "docker"

      env {
        CONSUL_HTTP_ADDR          = "172.17.0.1:8500"
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
        args       = ["-t", "1883", "--cluster-bind-port=3500", "--listener_gossip-bind-port=3100", "--listener-bind-port=4000"]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          mqtt    = 1883
          listener = 4000
          listener_gossip  = 3100
        }
      }

      resources {
        cpu    = 200
        memory = 32

        network {
          mbits = 10
          port  "mqtt"{}
          port  "cluster"{}
          port  "health"{}
          port  "listener"{}
          port  "listener_gossip"{}
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

    task "tcp-listener" {
      driver = "docker"

      env {
        TLS_CN           = "broker.iot.cloud.vx-labs.net"
        CONSUL_HTTP_ADDR = "$${NOMAD_IP_health}:8500"
        VAULT_ADDR       = "http://$${NOMAD_IP_health}:8200"
      }

      template {
        destination = "local/proxy.conf"
        env = true
        data = <<EOH
{{with secret "secret/data/vx/mqtt"}}
http_proxy="{{.Data.http_proxy}}"
https_proxy="{{.Data.http_proxy}}"
LE_EMAIL="{{.Data.acme_email}}"
no_proxy="10.0.0.0/8,172.16.0.0/12"
{{end}}
        EOH
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
        args       = ["-s", "8883", "--cluster-bind-port=3500", "--listener_gossip-bind-port=3100", "--listener-bind-port=4000"]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          listener = 4000
          listener_gossip  = 3100
          mqtts   = 8883
        }
      }

      resources {
        cpu    = 200
        memory = 32

        network {
          mbits = 10
          port  "mqtts"{}
          port  "cluster"{}
          port  "health"{}
          port  "listener"{}
          port  "listener_gossip"{}
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

    task "wss-listener" {
      driver = "docker"

      env {
        TLS_CN           = "broker.iot.cloud.vx-labs.net"
        CONSUL_HTTP_ADDR = "$${NOMAD_IP_health}:8500"
        VAULT_ADDR       = "http://$${NOMAD_IP_health}:8200"
      }

      template {
        destination = "local/proxy.conf"
        env = true
        data = <<EOH
{{with secret "secret/data/vx/mqtt"}}
http_proxy="{{.Data.http_proxy}}"
https_proxy="{{.Data.http_proxy}}"
LE_EMAIL="{{.Data.acme_email}}"
no_proxy="10.0.0.0/8,172.16.0.0/12"
{{end}}
        EOH
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
        args       = ["-w", "8008", "--cluster-bind-port=3500", "--listener_gossip-bind-port=3100", "--listener-bind-port=4000"]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          listener = 4000
          listener_gossip  = 3100
          wss     = 8008
        }
      }

      resources {
        cpu    = 200
        memory = 32

        network {
          mbits = 10
          port  "wss" {}
          port  "cluster"{}
          port  "health"{}
          port  "listener"{}
          port  "listener_gossip"{}
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
