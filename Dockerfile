# 1st stage
FROM golang:1-alpine AS builder
RUN apk --no-cache add git
WORKDIR /go/src/app
COPY *.go /go/src/app/
RUN go get -d -v ./...
RUN go install -v ./...


# 2nd stage
FROM alpine:3.9

RUN apk --no-cache add \
    docker \
    py-pip \
    python-dev \
    libffi-dev \
    openssl-dev \
    gcc \
    libc-dev \
    make \
    curl \
    && \
    pip install docker-compose

COPY --from=builder /go/bin/app /usr/local/bin/docker-compose-watcher

CMD ["docker-compose-watcher", "-once=0", "-printSettings", "-cleanup"]