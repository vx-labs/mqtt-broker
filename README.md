# MQTT Broker

Experiments around a gossip-based MQTT Broker.
Inspired (a lot) by https://emitter.io/.

## Running

The system is composed of multiple services, discovering themselves using a gossip-based service mesh.
Inter-service communication is based on GRPC.

### Single node

You can run 3 instances of the broker with the following command.
```
go run ./cli/allinone/main.go -j localhost:3302 -j localhost:3303 --cluster-bind-port 3301 -t 1883 &
go run ./cli/allinone/main.go -j localhost:3301 -j localhost:3303 --cluster-bind-port 3302 -t 1884 &
go run ./cli/allinone/main.go -j localhost:3302 -j localhost:3301 --cluster-bind-port 3303 -t 1885
```

The broker should start and listen on 0.0.0.0:1883 (tcp), 0.0.0.0:1884 (tcp) and 0.0.0.0:1885 (tcp).
You can connect to it using an MQTT client:

```
mosquitto_sub -t 'test' -d -q 1 &
mosquitto_pub --port 1884 -t 'test' -d -q 1 -m 'hello'
```