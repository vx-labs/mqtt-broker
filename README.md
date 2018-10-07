# MQTT Broker

Experiments around a gossip-based MQTT Broker.
Inspired (a lot) by https://emitter.io/.

## Running

The broker is released as a docker image.

```
docker run -p 1883:1883 --rm quay.io/vxlabs/mqtt-broker:v0.0.1 -t 1883
```

### Single node

You can start a single instance of the broker by running
```
go run ./cmd/broker/main.go  -t 1883
```

(`go run` can be replaced by the appropriate `docker run` command)

The broker should start and listen on 0.0.0.0:1883 (tcp).
You can connect to it using an MQTT client:

```
mosquitto_sub -t 'test' -d -q 1 &
mosquitto_pub -t 'test' -d -q 1 -m 'hello'
```

### Multiple nodes

Once one node is running, you can start other nodes and tell them to join the first node by running

```
go run ./cmd/broker/main.go  -t 1884 -j <first broker address>
go run ./cmd/broker/main.go  -t 1885 -j <first broker address>
```

You can find the join address to give in the first broker log messages.
