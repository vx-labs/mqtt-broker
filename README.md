# MQTT Broker

Experiments around a gossip-based MQTT Broker.
Inspired (a lot) by https://emitter.io/.

## Running

The system is composed of two services: a "broker" and a "listener".
The listener is listening on the network (tcp, tls, ws or wss), and forward MQTT packets
to the broker using a GRPC connection.

The broker is responsible of routing MQTT messages.

### Single node

You can run the two services by running:
```
docker run --net=host --rm quay.io/vxlabs/mqtt-listener:<tag> -t 1883
docker run --net=host --rm quay.io/vxlabs/mqtt-broker:<tag> -j localhost:3500 --cluster-bind-port=3501
```

(`docker run` can be replaced by the appropriate `go run` command, each entrypint is located
in the ./cmd/ folder)

The broker should start and listen on 0.0.0.0:1883 (tcp).
You can connect to it using an MQTT client:

```
mosquitto_sub -t 'test' -d -q 1 &
mosquitto_pub -t 'test' -d -q 1 -m 'hello'
```

### Multiple nodes

You can start any number of listener or broker, they will discover themselves using a gossip-based mesh network.