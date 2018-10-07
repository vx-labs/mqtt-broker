FROM quay.io/vxlabs/dep as builder

RUN mkdir -p $GOPATH/src/github.com/vx-labs
WORKDIR $GOPATH/src/github.com/vx-labs/mqtt-broker
RUN mkdir release
COPY Gopkg* ./
RUN dep ensure -vendor-only
COPY . ./
RUN go test ./... && \
    go build -buildmode=exe -ldflags="-s -w" -a -o /bin/broker ./cmd/broker

FROM alpine
EXPOSE 1883
ENTRYPOINT ["/usr/bin/server"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=builder /bin/broker /usr/bin/server
