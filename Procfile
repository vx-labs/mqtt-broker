node1: go run ./cli/allinone/main.go --pprof -j localhost:3302 -j localhost:3303 -j 3304 --cluster-bind-port 3301  -t 1883 --api-tcp-port 8081
node2: go run ./cli/allinone/main.go -j localhost:3301 -j localhost:3303 -j 3304 --cluster-bind-port 3302  -t 1884
node3: go run ./cli/allinone/main.go -j localhost:3302 -j localhost:3301 -j 3304  --cluster-bind-port 3303  -t 1885
topics: go run ./cli/topics/main.go -j localhost:3302 -j localhost:3301 -j localhost:3303 -j 3305 --cluster-bind-port 3304
router: go run ./cli/router/main.go -j localhost:3302 -j localhost:3301 -j localhost:3303 -j 3304 --cluster-bind-port 3305
