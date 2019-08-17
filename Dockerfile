FROM quay.io/vxlabs/dep as deps
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
RUN mkdir release
COPY Gopkg* ./
RUN dep ensure -vendor-only

FROM quay.io/vxlabs/dep as builder
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
RUN mkdir release
COPY --from=deps $GOPATH/src/github.com/vx-labs/mqtt-broker/vendor/ ./vendor/
COPY . ./
RUN go test ./...

FROM quay.io/vxlabs/dep as api-builder
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
RUN mkdir release
COPY --from=deps $GOPATH/src/github.com/vx-labs/mqtt-broker/vendor/ ./vendor/
COPY . ./
RUN go build -buildmode=exe -ldflags="-s -w" -a -o /bin/api ./cli/api

FROM quay.io/vxlabs/dep as broker-builder
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
RUN mkdir release
COPY --from=deps $GOPATH/src/github.com/vx-labs/mqtt-broker/vendor/ ./vendor/
COPY . ./
RUN go build -buildmode=exe -ldflags="-s -w" -a -o /bin/broker ./cli/broker

FROM quay.io/vxlabs/dep as listener-builder
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
RUN mkdir release
COPY --from=deps $GOPATH/src/github.com/vx-labs/mqtt-broker/vendor/ ./vendor/
COPY . ./
RUN go build -buildmode=exe -ldflags="-s -w" -a -o /bin/listener ./cli/listener

FROM quay.io/vxlabs/dep as sessions-builder
RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
RUN mkdir release
COPY --from=deps $GOPATH/src/github.com/vx-labs/mqtt-broker/vendor/ ./vendor/
COPY . ./
RUN go build -buildmode=exe -ldflags="-s -w" -a -o /bin/sessions ./cli/sessions

FROM alpine as broker
EXPOSE 1883
ENTRYPOINT ["/usr/bin/server"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=broker-builder /bin/broker /usr/bin/server

FROM alpine as api
EXPOSE 1883
ENTRYPOINT ["/usr/bin/server"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=api-builder /bin/api /usr/bin/server

FROM alpine as listener
EXPOSE 1883
ENTRYPOINT ["/usr/bin/server"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=listener-builder /bin/listener /usr/bin/server


FROM alpine as sessions
EXPOSE 1883
ENTRYPOINT ["/usr/bin/server"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=sessions-builder /bin/sessions /usr/bin/server

