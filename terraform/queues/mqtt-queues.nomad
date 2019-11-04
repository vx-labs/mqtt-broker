job "mqtt-queues" {
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

  group "queues" {
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
        CONSUL_HTTP_ADDR          = "172.17.0.1:8500"
      }

      config {
        logging {
          type = "fluentd"

          config {
            fluentd-address = "localhost:24224"
            tag             = "mqtt-queues"
          }
        }

        image      = "quay.io/vxlabs/mqtt-queues:${broker_version}"
        args       = [
          "--cluster-bind-port=3500",
          "--queues_gossip-bind-port=3100",
          "--queues-bind-port=4000",
          "--queues_gossip_rpc-bind-port=3200",
        ]
        force_pull = true

        port_map {
          health  = 9000
          cluster = 3500
          queues = 4000
          queues_gossip  = 3100
          queues_gossip_rpc  = 3200
        }
      }

      resources {
        cpu    = 200
        memory = 128

        network {
          mbits = 10
          port  "cluster"{}
          port  "health"{}
          port  "queues"{}
          port  "queues_gossip"{}
          port  "queues_gossip_rpc"{}
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
