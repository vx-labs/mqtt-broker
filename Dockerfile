
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

FROM builder as subscriptions-builder
ARG BUILT_VERSION="n/a"
RUN go build -buildmode=exe -ldflags="-s -w -X github.com/vx-labs/mqtt-broker/cli.BuiltVersion=${BUILT_VERSION}" -a -o /bin/subscriptions ./cli/subscriptions

FROM builder as api-builder
ARG BUILT_VERSION="n/a"
RUN go build -buildmode=exe -ldflags="-s -w -X github.com/vx-labs/mqtt-broker/cli.BuiltVersion=${BUILT_VERSION}" -a -o /bin/api ./cli/api

FROM builder as broker-builder
ARG BUILT_VERSION="n/a"
RUN go build -buildmode=exe -ldflags="-s -w -X github.com/vx-labs/mqtt-broker/cli.BuiltVersion=${BUILT_VERSION}" -a -o /bin/broker ./cli/broker

FROM builder as listener-builder
ARG BUILT_VERSION="n/a"
RUN go build -buildmode=exe -ldflags="-s -w -X github.com/vx-labs/mqtt-broker/cli.BuiltVersion=${BUILT_VERSION}" -a -o /bin/listener ./cli/listener

FROM builder as sessions-builder
ARG BUILT_VERSION="n/a"
RUN go build -buildmode=exe -ldflags="-s -w -X github.com/vx-labs/mqtt-broker/cli.BuiltVersion=${BUILT_VERSION}" -a -o /bin/sessions ./cli/sessions

FROM alpine as prod
ENTRYPOINT ["/usr/bin/server"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*

FROM prod as broker
COPY --from=broker-builder /bin/broker /usr/bin/server

FROM prod as subscriptions
COPY --from=subscriptions-builder /bin/subscriptions /usr/bin/server

FROM prod as api
COPY --from=api-builder /bin/api /usr/bin/server

FROM prod as listener
COPY --from=listener-builder /bin/listener /usr/bin/server

FROM prod as sessions
COPY --from=sessions-builder /bin/sessions /usr/bin/server

