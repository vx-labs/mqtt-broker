# MQTT Broker

Experiments around a gossip-based MQTT Broker.
Inspired (initially) by https://emitter.io/.

## Running

The broker is composed of multiple services, discovering themselves using an embed gossip-based service mesh.

Inter-service communication is based on GRPC.

The project use Goreman (https://github.com/mattn/goreman/) to run on a local development workstation.

### Local

The system persists its state on disk. You may have to delete the persited state before starting the system.

Persisted state is located in the ~/.local/share/mqtt-broker folder (or /var/lib/mqtt-broker if you run the system as the root user).

You can run the broker on your workstation by running the following command.
```
goreman start
```

Wait a few seconds for raft clusters (key-value store, stream service and queues service) to bootstrap themselves.

The broker should start and listen on 0.0.0.0:1883 (tcp) and 0.0.0.0:1884 (tcp).
You can connect to it using an MQTT client like mosquitto.

```
mosquitto_sub -t 'test' -d -q 1 &
mosquitto_pub --port 1884 -t 'test' -d -q 1 -m 'hello'
```
