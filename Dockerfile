ARG ALPINE_VERSION=3.13
ARG GO_VERSION=1.16
ARG COMPOSE_VERSION=1.28.6

FROM amd64/golang:${GO_VERSION}-alpine AS builder
RUN apk --update add --no-cache git
RUN export GOBIN=$HOME/work/bin
WORKDIR /go/src/app
ADD src/ .
ADD src/go.mod .
ADD src/go.sum .
RUN go get -d -v ./...
RUN CGO_ENABLED=1 go build -o main .

FROM amd64/alpine:${ALPINE_VERSION}
ARG COMPOSE_VERSION
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
        org.label-schema.name="Compose Updater" \
        org.label-schema.description="Automatically update your Docker Compose containers." \
        org.label-schema.vcs-ref=$VCS_REF \
        org.label-schema.vcs-url="https://github.com/virtualzone/compose-updater" \
        org.label-schema.schema-version="1.0"
RUN apk --no-cache add \
    docker \
    python3 \
    py3-requests \
    py3-distro
RUN apk --no-cache --virtual .build-deps add \
    py3-pip \
    python3-dev \
    libffi-dev \
    openssl-dev \
    gcc \
    libc-dev \
    make \
    rust \
    cargo \
    && pip3 install --upgrade pip \
    && pip3 install --no-cache-dir docker-compose==$COMPOSE_VERSION \
    && apk del .build-deps
COPY --from=builder /go/src/app/main /usr/local/bin/docker-compose-watcher
CMD ["docker-compose-watcher", "-once=0", "-printSettings", "-cleanup"]