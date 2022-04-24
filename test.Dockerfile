ARG ALPINE_VERSION=3.15
ARG GO_VERSION=1.18
ARG COMPOSE_VERSION=2.4.1

FROM golang:${GO_VERSION} AS builder
RUN export GOBIN=$HOME/work/bin
WORKDIR /go/src/app
ADD src/ .
ADD src/go.mod .
ADD src/go.sum .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main .

FROM alpine:${ALPINE_VERSION}
ARG COMPOSE_VERSION
ARG TARGETPLATFORM
RUN apk --no-cache add docker curl bash
RUN \
    case ${TARGETPLATFORM} in \
      "linux/amd64")  DOWNLOAD_ARCH="x86_64"  ;; \
      "linux/arm64") DOWNLOAD_ARCH="aarch64"  ;; \
      "linux/arm/v7") DOWNLOAD_ARCH="armv7"  ;; \
      *) DOWNLOAD_ARCH="x86_64"  ;; \
    esac && \
    mkdir -p /usr/local/lib/docker/cli-plugins && \
    echo "https://github.com/docker/compose/releases/download/v${COMPOSE_VERSION}/docker-compose-linux-${DOWNLOAD_ARCH}" && \
    curl -SLf https://github.com/docker/compose/releases/download/v${COMPOSE_VERSION}/docker-compose-linux-${DOWNLOAD_ARCH} -o /usr/local/lib/docker/cli-plugins/docker-compose && \
    chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
WORKDIR /test
ADD ./test/ /test/
COPY --from=builder /go/src/app/main /test/docker-compose-watcher
EXPOSE 1883
CMD ["/test/run-test.sh"]