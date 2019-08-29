
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

FROM builder as binary-builder
ARG ARTIFACT="###"
RUN apk -U add git
RUN BUILD_VERSION=$(git rev-parse --short HEAD) && go build -buildmode=exe -ldflags="-s -w -X github.com/vx-labs/mqtt-broker/cli.BuiltVersion=${BUILT_VERSION}" -a -o /bin/${ARTIFACT} ./cli/${ARTIFACT}

FROM alpine as prod
ARG ARTIFACT="###"
ENTRYPOINT ["/usr/bin/server"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=binary-builder /bin/${ARTIFACT} /usr/bin/server
