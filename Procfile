api: go run ./cli/broker service api --cluster-bind-port 3500

messages.1: go run ./cli/broker service messages --cluster-bind-port $PORT -j localhost:3500 --initial-stream-config messages:1 --initial-stream-config events:1
messages.2: go run ./cli/broker service messages --cluster-bind-port $PORT -j localhost:3500 --initial-stream-config messages:1 --initial-stream-config events:1
messages.3: go run ./cli/broker service messages --cluster-bind-port $PORT -j localhost:3500 --initial-stream-config messages:1 --initial-stream-config events:1

kv.1: go run ./cli/broker service kv --cluster-bind-port $PORT -j localhost:3500
kv.2: go run ./cli/broker service kv --cluster-bind-port $PORT -j localhost:3500
kv.3: go run ./cli/broker service kv --cluster-bind-port $PORT -j localhost:3500

queues.1: go run ./cli/broker service queues --cluster-bind-port $PORT -j localhost:3500
queues.2: go run ./cli/broker service queues --cluster-bind-port $PORT -j localhost:3500
queues.3: go run ./cli/broker service queues --cluster-bind-port $PORT -j localhost:3500

sessions: go run ./cli/broker service sessions --cluster-bind-port $PORT -j localhost:3500
subscriptions: go run ./cli/broker service subscriptions --cluster-bind-port $PORT -j localhost:3500
router: go run ./cli/broker service router --cluster-bind-port $PORT -j localhost:3500
topics: go run ./cli/broker service topics --cluster-bind-port $PORT -j localhost:3500
broker: go run ./cli/broker service broker --cluster-bind-port $PORT -j localhost:3500

listener.1: go run ./cli/broker service listener --cluster-bind-port $PORT -j localhost:3500 --listener-tcp-port 1883
listener.2: go run ./cli/broker service listener --cluster-bind-port $PORT -j localhost:3500 --listener-tcp-port 1884
