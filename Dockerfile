ARG ALPINE_VERSION=3.9

FROM golang:1-alpine AS go_builder
RUN apk --no-cache add git
WORKDIR /go/src/app
COPY *.go /go/src/app/
RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:${ALPINE_VERSION}
RUN apk --no-cache add \
    docker \
    python2
RUN apk --no-cache --virtual .build-deps add \
    py-pip \
    python-dev \
    libffi-dev \
    openssl-dev \
    gcc \
    libc-dev \
    make \
    && pip install --no-cache-dir docker-compose \
    && apk del .build-deps
COPY --from=go_builder /go/bin/app /usr/local/bin/docker-compose-watcher
CMD ["docker-compose-watcher", "-once=0", "-printSettings", "-cleanup"]