
FROM quay.io/vxlabs/dep as deps
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
COPY Gopkg* ./
RUN dep ensure -vendor-only

FROM quay.io/vxlabs/dep as builder
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
COPY --from=deps $GOPATH/src/github.com/vx-labs/mqtt-broker/vendor/ ./vendor/
COPY . ./
RUN go test ./...
ARG BUILT_VERSION="snapshot"
RUN go build -buildmode=exe -ldflags="-s -w -X github.com/vx-labs/mqtt-broker/cli.BuiltVersion=${BUILT_VERSION}" \
       -a -o /bin/broker ./cli/broker

FROM alpine as prod
ENTRYPOINT ["/usr/bin/broker"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=builder /bin/broker /usr/bin/broker
