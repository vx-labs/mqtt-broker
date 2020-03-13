api: go run ./cli/broker service api --cluster-bind-port 3500 --api-tcp-port 8081 --node-id api

messages-1: go run ./cli/broker service messages -j localhost:3500 --initial-stream-config messages:1 --initial-stream-config events:1 --node-id messages-1
messages-2: go run ./cli/broker service messages -j localhost:3500 --initial-stream-config messages:1 --initial-stream-config events:1 --node-id messages-2
messages-3: go run ./cli/broker service messages -j localhost:3500 --initial-stream-config messages:1 --initial-stream-config events:1 --node-id messages-3

kv-1: go run ./cli/broker service kv -j localhost:3500 --node-id kv-1
kv-2: go run ./cli/broker service kv -j localhost:3500 --node-id kv-2
kv-3: go run ./cli/broker service kv -j localhost:3500 --node-id kv-3

queues-1: go run ./cli/broker service queues -j localhost:3500 --node-id queues-1
queues-2: go run ./cli/broker service queues -j localhost:3500 --node-id queues-2
queues-3: go run ./cli/broker service queues -j localhost:3500 --node-id queues-3

sessions: go run ./cli/broker service sessions -j localhost:3500 --node-id sessions
subscriptions: go run ./cli/broker service subscriptions -j localhost:3500 --node-id subscriptions
router: go run ./cli/broker service router -j localhost:3500 --node-id router
topics: go run ./cli/broker service topics -j localhost:3500 --node-id topics
broker: go run ./cli/broker service broker -j localhost:3500 --node-id broker
auth: go run ./cli/broker service auth -j localhost:3500 --node-id auth

listener-1: go run ./cli/broker service listener -j localhost:3500 --listener-tcp-port 1883 --node-id listener-1
listener-2: go run ./cli/broker service listener -j localhost:3500 --listener-tcp-port 1884 --node-id listener-2
